// c-flat-uds: FLAT-architecture HMAC-SHA256 JWT verifier over UDS.
//
// One file, no abstractions, all inline.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
#include <signal.h>
#include <sys/stat.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <openssl/hmac.h>
#include <openssl/evp.h>
#include <openssl/sha.h>
#include <openssl/bio.h>
#include <openssl/buffer.h>

#define MAX_LINE 2048
#define SECRET_MAX 256

static char g_secret[SECRET_MAX] = {0};
static int g_secret_len = 0;

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

static void handle(int fd) {
    char buf[MAX_LINE];
    while (1) {
        ssize_t n = read(fd, buf, sizeof(buf) - 1);
        if (n <= 0) break;
        buf[n] = 0;
        char *nl = strchr(buf, '\n');
        if (nl) *nl = 0;
        const char *resp;
        char okbuf[256];
        if (strncmp(buf, "VERIFY ", 7) != 0) {
            resp = "ERR BAD_REQUEST\n";
            ssize_t w = write(fd, resp, strlen(resp)); (void)w;
            continue;
        }
        char sub[128] = {0};
        int r = verify_jwt(buf + 7, sub, sizeof(sub));
        if (r == 0) {
            snprintf(okbuf, sizeof(okbuf), "OK %s\n", sub);
        } else if (r == -2) {
            snprintf(okbuf, sizeof(okbuf), "ERR BAD_REQUEST\n");
        } else {
            snprintf(okbuf, sizeof(okbuf), "ERR INVALID_TOKEN\n");
        }
        ssize_t w = write(fd, okbuf, strlen(okbuf)); (void)w;
    }
    close(fd);
}

int main(int argc, char **argv) {
    (void)argc; (void)argv;
    const char *secret = getenv("JWT_SECRET");
    const char *path_env = getenv("UDS_PATH");
    if (!secret) { fprintf(stderr, "JWT_SECRET required\n"); return 1; }
    const char *path = path_env ? path_env : "/tmp/c-flat-uds.sock";
    g_secret_len = strlen(secret);
    if (g_secret_len >= SECRET_MAX) { fprintf(stderr, "secret too long\n"); return 1; }
    memcpy(g_secret, secret, g_secret_len);

    unlink(path);
    int srv = socket(AF_UNIX, SOCK_STREAM, 0);
    struct sockaddr_un sa = {0};
    sa.sun_family = AF_UNIX;
    strncpy(sa.sun_path, path, sizeof(sa.sun_path) - 1);
    if (bind(srv, (struct sockaddr *)&sa, sizeof(sa)) < 0) { perror("bind"); return 1; }
    listen(srv, 64);
    chmod(path, 0660);
    fprintf(stderr, "c-flat-uds listening on %s\n", path);
    signal(SIGCHLD, SIG_IGN);
    while (1) {
        int fd = accept(srv, NULL, NULL);
        if (fd < 0) continue;
        if (fork() == 0) { close(srv); handle(fd); return 0; }
        close(fd);
    }
    return 0;
}
