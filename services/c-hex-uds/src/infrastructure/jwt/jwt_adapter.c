// c-hex-http/src/infrastructure/jwt: the OUTGOING ADAPTER.
#include "jwt_adapter.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <openssl/hmac.h>
#include <openssl/evp.h>
#include <openssl/sha.h>
#include <openssl/bio.h>
#include <openssl/buffer.h>

char g_secret[AB_SECRET_MAX] = {0};
int g_secret_len = 0;

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

int jwt_adapter_verify(const char *token, char *out_sub, int sub_sz, int *out_err) {
    if (out_err) *out_err = -1;
    out_sub[0] = 0;
    if (!token || !*token) {
        if (out_err) *out_err = -2;
        return -2;
    }
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
    if (out_err) *out_err = 0;
    return 0;
}

void jwt_adapter_port(TokenVerifierPort *out) {
    out->verify = jwt_adapter_verify;
}
