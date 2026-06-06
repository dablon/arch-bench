// c-hex-http/src/infrastructure/jwt: the OUTGOING ADAPTER. Implements
// the domain's TokenVerifierPort using HMAC-SHA256 JWT. The whole
// HMAC, base64url, JSON, exp-checking pipeline lives here and ONLY
// here.
#ifndef AB_C_HEX_JWT_ADAPTER_H
#define AB_C_HEX_JWT_ADAPTER_H

#include "../../domain/port.h"
#define AB_SECRET_MAX 256

extern char g_secret[AB_SECRET_MAX];
extern int g_secret_len;

int jwt_adapter_verify(const char *token, char *out_sub, int sub_sz, int *out_err);

// Construct a TokenVerifierPort that dispatches to jwt_adapter_verify.
void jwt_adapter_port(TokenVerifierPort *out);

#endif
