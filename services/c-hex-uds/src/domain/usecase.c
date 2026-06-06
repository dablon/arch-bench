// c-hex-http/src/domain/usecase.c: the use case implementation.
#include "usecase.h"
#include <string.h>

void usecase_verify(const TokenVerifierPort *port, const char *token, VerificationResult *out) {
    out->valid = 0;
    out->subject[0] = 0;
    out->code = DCODE_INVALID_TOKEN;

    if (!token || !*token) {
        out->code = DCODE_BAD_REQUEST;
        return;
    }
    if (!port || !port->verify) {
        out->code = DCODE_INVALID_TOKEN;
        return;
    }
    int err = 0;
    port->verify(token, out->subject, sizeof(out->subject), &err);
    if (err == 0) {
        out->valid = 1;
        out->code = DCODE_OK;
    } else if (err == -2) {
        out->code = DCODE_BAD_REQUEST;
    } else {
        out->code = DCODE_INVALID_TOKEN;
    }
}
