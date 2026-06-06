// Package usecase contains the application's business rules.
//
// A use case is the smallest unit of "what the system does". For a
// verifier service the only use case is "verify a token". The use case:
//   - receives a domain input (a token string)
//   - calls the outgoing port it depends on
//   - returns a domain result
//
// Use cases are framework-free. They don't know about HTTP, JSON,
// gRPC, JWT, or any specific crypto algorithm.
package usecase

import (
	"github.com/dablon/arch-bench/services/go-hex-http/domain/entity"
	"github.com/dablon/arch-bench/services/go-hex-http/domain/port"
)

// VerifyToken is the canonical use case. It is the only thing the
// application does.
//
// Inputs and outputs are domain types. The adapter (infrastructure/
// adapter/http) is responsible for translating transport-level types
// (JSON bodies, gRPC requests) into the input shape, and the
// translation of the output into the response.
type VerifyToken struct {
	tokens port.TokenVerifier
}

func NewVerifyToken(t port.TokenVerifier) *VerifyToken {
	return &VerifyToken{tokens: t}
}

// Execute runs the use case. Token "" is rejected at the use-case
// level — it is a business rule, not a transport rule.
func (u *VerifyToken) Execute(token string) entity.VerificationResult {
	if token == "" {
		return entity.VerificationResult{Code: entity.CodeBadRequest}
	}
	sub, err := u.tokens.Verify(token)
	if err != nil {
		return entity.VerificationResult{Code: entity.CodeInvalidToken}
	}
	return entity.VerificationResult{Valid: true, Subject: sub, Code: entity.CodeOK}
}
