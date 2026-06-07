// Package usecase holds the application's business rules.
package usecase

import (
	"github.com/dablon/arch-bench/services/go-hex-uds/domain/entity"
	"github.com/dablon/arch-bench/services/go-hex-uds/domain/port"
)

type VerifyToken struct{ tokens port.TokenVerifier }

func NewVerifyToken(t port.TokenVerifier) *VerifyToken {
	return &VerifyToken{tokens: t}
}

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
