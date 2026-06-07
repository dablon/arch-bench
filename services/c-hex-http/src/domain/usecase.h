// c-hex-http/src/domain: the USE CASE. The only piece of business
// logic the application has. Calls the port, maps errors to domain
// codes, returns a VerificationResult.
#ifndef AB_C_HEX_DOMAIN_USECASE_H
#define AB_C_HEX_DOMAIN_USECASE_H

#include "result.h"
#include "port.h"

void usecase_verify(const TokenVerifierPort *port, const char *token, VerificationResult *out);

#endif
