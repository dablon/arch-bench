// c-hex-http/src/domain: the pure-domain value types. No OpenSSL, no
// HTTP, no JSON, no socket headers. Only the "what does the system
// mean" types.
#ifndef AB_C_HEX_DOMAIN_RESULT_H
#define AB_C_HEX_DOMAIN_RESULT_H

typedef enum {
    DCODE_OK = 0,
    DCODE_BAD_REQUEST = 1,
    DCODE_INVALID_TOKEN = 2,
} DomainCode;

typedef struct {
    int valid;
    char subject[128];
    DomainCode code;
} VerificationResult;

#endif
