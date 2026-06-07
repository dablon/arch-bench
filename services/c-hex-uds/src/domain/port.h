// c-hex-http/src/domain: the OUTGOING PORT. The use case depends on
// this interface; an adapter in infrastructure/ implements it.
//
// This is the dependency rule: domain/ has no idea what the JWT
// library is. The use case calls verify_token on a function pointer
// supplied at composition time.
#ifndef AB_C_HEX_DOMAIN_PORT_H
#define AB_C_HEX_DOMAIN_PORT_H

#include "result.h"

typedef struct {
    // Adapter sets these at construction.
    int (*verify)(const char *token, char *out_sub, int sub_sz, int *out_err);
} TokenVerifierPort;

#endif
