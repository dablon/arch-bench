#include "token_service.h"
#include "../repo/jwt_repo.h"

void service_verify(const char *token, ServiceResult *out) {
    out->valid = 0;
    out->subject[0] = 0;
    out->code = CODE_INVALID_TOKEN;

    if (!token || !*token) {
        out->code = CODE_BAD_REQUEST;
        return;
    }
    int r = repo_verify(token, out->subject, sizeof(out->subject));
    if (r == 0) {
        out->valid = 1;
        out->code = CODE_OK;
    } else if (r == -2) {
        out->code = CODE_BAD_REQUEST;
    } else {
        out->code = CODE_INVALID_TOKEN;
    }
}
