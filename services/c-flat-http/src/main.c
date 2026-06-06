// c-flat-http: FLAT-architecture HMAC-SHA256 JWT verifier over HTTP.
//
// Architectural style: FLAT. ONE file. Every concern — JSON parsing,
// HTTP framing, JWT decoding, signature verification, base64url,
// JSON response, env config — lives in this single .c file. No
// separate .h, no shared object, no static library. The reader can
// see the full request-to-response pipeline in 350 lines of straight
// C.
//
// When this grows past ~700 lines (multiple endpoints, multiple token
// formats, tests, observability, multiple transports) the lack of
// seams hurts. For a single-purpose service like a verifier, flat is
// honest.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
#include <signal.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <openssl/hmac.h>
#include <openssl/evp.h>
#include <openssl/sha.h>
#include <openssl/bio.h>
#include <openssl/buffer.h>

#define MAX_REQ 8192
#define SECRET_MAX 256

static char g_secret[SECRET_MAX];
static int g_secret_len = 0;

/* === base64url helpers =========================================== */
static int b64url_decode(const char *in, int inlen, unsigned char *out, int *outlen) {
    char tmp[1024];
    int pad = (4 - (inlen % 4)) % 4;
    if (inlen + pad >= (int)sizeof(tmp)) return -1;
    for (int i = 0; i < inlen; i++) {
        char c = in[i];
        if (c == '-') c = '+';
        else if (c == '_') c = '/';
        tmp[i] = c;
    }
    for (int i = 0; i < pad; i++) tmp[inlen + i] = '=';
    tmp[inlen + pad] = 0;
    BIO *b = BIO_new_mem_buf(tmp, inlen + pad);
    BIO *b64 = BIO_new(BIO_f_base64());
    BIO_set_flags(b64, BIO_FLAGS_BASE64_NO_NL);
    b = BIO_push(b64, b);
    int n = BIO_read(b, out, inlen + pad);
    BIO_free_all(b);
    if (n <= 0) return -1;
    *outlen = (int)n;
    return 0;
}

static int b64url_encode(const unsigned char *in, int inlen, char *out, int outsz) {
    BIO *b64 = BIO_new(BIO_f_base64());
    BIO_set_flags(b64, BIO_FLAGS_BASE64_NO_NL);
    BIO *bmem = BIO_new(BIO_s_mem());
    b64 = BIO_push(b64, bmem);
    BIO_write(b64, in, inlen);
    (void)BIO_flush(b64);
    BUF_MEM *bptr;
    BIO_get_mem_ptr(b64, &bptr);
    int n = bptr->length;
    if (n >= outsz) { BIO_free_all(b64); return -1; }
    memcpy(out, bptr->data, n);
    out[n] = 0;
    BIO_free_all(b64);
    for (int i = 0; i < n; i++) {
        if (out[i] == '+') out[i] = '-';
        else if (out[i] == '/') out[i] = '_';
        else if (out[i] == '=') { out[i] = 0; n = i; break; }
    }
    return n;
}

static int extract_b64(const char *seg, int seglen, char *out, int outsz) {
    int n = 0;
    for (int i = 0; i < seglen && n < outsz - 1; i++) {
        char c = seg[i];
        if (c == '.') break;
        if (c == '=') continue;
        out[n++] = c;
    }
    out[n] = 0;
    return n;
}

/* === tiny JSON parser ============================================= */
static int json_get_str(const char *json, const char *key, char *out, int outsz) {
    char needle[64];
    snprintf(needle, sizeof(needle), "\"%s\":\"", key);
    const char *p = strstr(json, needle);
    if (!p) return -1;
    p += strlen(needle);
    const char *e = strchr(p, '"');
    if (!e) return -1;
    int len = e - p;
    if (len >= outsz) len = outsz - 1;
    memcpy(out, p, len);
    out[len] = 0;
    return 0;
}

static int json_get_num(const char *json, const char *key, double *out) {
    char needle[64];
    snprintf(needle, sizeof(needle), "\"%s\":", key);
    const char *p = strstr(json, needle);
    if (!p) return -1;
    p += strlen(needle);
    *out = strtod(p, NULL);
    return 0;
}

/* === JWT verify =================================================== */
/* Returns: 0=ok, -1=invalid, -2=bad request */
static int verify_jwt(const char *token, char *out_sub, int sub_sz) {
    out_sub[0] = 0;
    if (!token || !*token) return -2;
    const char *p1 = strchr(token, '.');
    if (!p1) return -1;
    const char *p2 = strchr(p1 + 1, '.');
    if (!p2) return -1;
    int header_len = p1 - token;
    int payload_len = p2 - (p1 + 1);
    int signing_input_len = header_len + 1 + payload_len;
    char *signing_input = malloc(signing_input_len + 1);
    memcpy(signing_input, token, signing_input_len);
    signing_input[signing_input_len] = 0;
    unsigned char mac[EVP_MAX_MD_SIZE];
    unsigned int maclen = 0;
    unsigned char *r = HMAC(EVP_sha256(), g_secret, g_secret_len, (unsigned char *)signing_input, signing_input_len, mac, &maclen);
    free(signing_input);
    if (!r) return -1;
    char expected[128];
    int explen = b64url_encode(mac, maclen, expected, sizeof(expected));
    if (explen < 0) return -1;
    char provided[128];
    int provlen = extract_b64(p2 + 1, strlen(p2 + 1), provided, sizeof(provided));
    if (provlen != explen) return -1;
    unsigned int diff = 0;
    for (int i = 0; i < explen; i++) diff |= (unsigned char)expected[i] ^ (unsigned char)provided[i];
    if (diff != 0) return -1;
    char payload[1024];
    if (extract_b64(p1 + 1, payload_len, payload, sizeof(payload)) == 0) return -1;
    unsigned char decoded[1024];
    int declen = 0;
    if (b64url_decode(payload, (int)strlen(payload), decoded, &declen) != 0) return -1;
    decoded[declen] = 0;
    if (json_get_str((char *)decoded, "sub", out_sub, sub_sz) != 0) {
        out_sub[0] = 0;
    }
    double exp_val = 0;
    if (json_get_num((char *)decoded, "exp", &exp_val) == 0) {
        time_t now = time(NULL);
        if ((double)now > exp_val) return -1;
    }
    return 0;
}

/* === HTTP transport =============================================== */
static void send_response(int fd, int code, const char *body) {
    const char *status = (code == 200) ? "OK" : (code == 400) ? "Bad Request" : (code == 401) ? "Unauthorized" : (code == 404) ? "Not Found" : (code == 405) ? "Method Not Allowed" : "Error";
    char resp[1024];
    int n = snprintf(resp, sizeof(resp),
        "HTTP/1.1 %d %s\r\nContent-Type: application/json\r\nContent-Length: %zu\r\nConnection: close\r\n\r\n%s",
        code, status, strlen(body), body);
    if (n > 0) { ssize_t _w = write(fd, resp, n); (void)_w; }
}

static int read_http(int fd, char *buf, int sz) {
    int total = 0;
    while (total < sz - 1) {
        int n = read(fd, buf + total, sz - 1 - total);
        if (n <= 0) break;
        total += n;
        buf[total] = 0;
        if (strstr(buf, "\r\n\r\n")) break;
    }
    return total;
}

static void handle(int fd) {
    char buf[MAX_REQ];
    int n = read_http(fd, buf, sizeof(buf));
    if (n <= 0) { close(fd); return; }
    char method[16] = {0}, path[256] = {0};
    sscanf(buf, "%15s %255s", method, path);
    if (strcmp(path, "/health") == 0) {
        send_response(fd, 200, "{\"status\":\"ok\"}");
    } else if (strcmp(path, "/verify") == 0) {
        if (strcmp(method, "POST") != 0) {
            send_response(fd, 405, "{\"valid\":false,\"code\":\"ERR_METHOD\"}");
        } else {
            char *body = strstr(buf, "\r\n\r\n");
            if (!body) { send_response(fd, 400, "{\"valid\":false,\"code\":\"ERR_BAD_REQUEST\"}"); close(fd); return; }
            body += 4;
            char token[1024] = {0};
            if (json_get_str(body, "token", token, sizeof(token)) != 0 || !*token) {
                send_response(fd, 400, "{\"valid\":false,\"code\":\"ERR_BAD_REQUEST\"}");
            } else {
                char sub[128] = {0};
                int r = verify_jwt(token, sub, sizeof(sub));
                if (r == 0) {
                    char resp[256];
                    snprintf(resp, sizeof(resp), "{\"valid\":true,\"subject\":\"%s\",\"code\":\"OK\"}", sub);
                    send_response(fd, 200, resp);
                } else if (r == -2) {
                    send_response(fd, 400, "{\"valid\":false,\"code\":\"ERR_BAD_REQUEST\"}");
                } else {
                    send_response(fd, 401, "{\"valid\":false,\"code\":\"ERR_INVALID_TOKEN\"}");
                }
            }
        }
    } else {
        send_response(fd, 404, "{\"valid\":false,\"code\":\"ERR_NOT_FOUND\"}");
    }
    close(fd);
}

int main(int argc, char **argv) {
    (void)argc; (void)argv;
    const char *secret = getenv("JWT_SECRET");
    const char *addr_env = getenv("LISTEN_ADDR");
    if (!secret) { fprintf(stderr, "JWT_SECRET required\n"); return 1; }
    int port = 9000;
    if (addr_env) {
        char *colon = strrchr(addr_env, ':');
        if (colon) port = atoi(colon + 1);
    }
    g_secret_len = strlen(secret);
    if (g_secret_len >= SECRET_MAX) { fprintf(stderr, "secret too long\n"); return 1; }
    memcpy(g_secret, secret, g_secret_len);

    int srv = socket(AF_INET, SOCK_STREAM, 0);
    int opt = 1;
    setsockopt(srv, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
    struct sockaddr_in sa = {0};
    sa.sin_family = AF_INET;
    sa.sin_addr.s_addr = htonl(INADDR_LOOPBACK);
    sa.sin_port = htons(port);
    if (bind(srv, (struct sockaddr *)&sa, sizeof(sa)) < 0) { perror("bind"); return 1; }
    listen(srv, 64);
    fprintf(stderr, "c-flat-http listening on 127.0.0.1:%d\n", port);
    signal(SIGCHLD, SIG_IGN);
    while (1) {
        int fd = accept(srv, NULL, NULL);
        if (fd < 0) continue;
        if (fork() == 0) { close(srv); handle(fd); return 0; }
        close(fd);
    }
    return 0;
}
