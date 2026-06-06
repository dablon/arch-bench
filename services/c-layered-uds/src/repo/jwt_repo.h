// c-layered-http/src/repo: data access. The only place that knows
// about HMAC-SHA256 JWT internals. Other layers use the TokenRepo port.
#ifndef AB_C_LAYERED_REPO_H
#define AB_C_LAYERED_REPO_H

#define AB_SECRET_MAX 256

extern char g_secret[AB_SECRET_MAX];
extern int g_secret_len;

// Returns: 0 = ok, -1 = invalid, -2 = bad request
int repo_verify(const char *token, char *out_sub, int sub_sz);

#endif
