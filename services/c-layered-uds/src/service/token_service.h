// c-layered-http/src/service: business logic. Maps repo return codes
// to domain result codes. The transport layer above never sees the
// repo's internal -1/-2; it sees CodeOK / CodeBadRequest /
// CodeInvalidToken.
#ifndef AB_C_LAYERED_SERVICE_H
#define AB_C_LAYERED_SERVICE_H

typedef enum {
    CODE_OK = 0,
    CODE_BAD_REQUEST = 1,
    CODE_INVALID_TOKEN = 2,
} ResultCode;

typedef struct {
    int valid;
    char subject[128];
    ResultCode code;
} ServiceResult;

// service_verify is the only function the service exposes. It calls
// the repo, translates the return, and returns a domain result.
void service_verify(const char *token, ServiceResult *out);

#endif
