// Package service is the business-logic layer of the layered architecture.
//
// The service layer:
//   - takes a domain-level input (a token string)
//   - delegates the actual crypto to the repository
//   - returns a domain-level result (a code like "OK" or "ERR_INVALID_TOKEN")
//
// The service knows nothing about HTTP, JSON, or any transport. The
// handler knows nothing about JWT, claims, or any crypto.
package service

import (
	"errors"

	"github.com/dablon/arch-bench/services/go-layered-http/internal/repo"
)

// ResultCode is the domain-level outcome of a verification. These
// values are stable across transports.
type ResultCode string

const (
	CodeOK            ResultCode = "OK"
	CodeBadRequest    ResultCode = "ERR_BAD_REQUEST"
	CodeInvalidToken  ResultCode = "ERR_INVALID_TOKEN"
)

// Result is the domain-level outcome. The service returns this to the
// handler; the handler maps it to whatever the transport wants.
type Result struct {
	Valid   bool
	Subject string
	Code    ResultCode
}

// VerifierService is the port the handler layer depends on.
type VerifierService interface {
	VerifyToken(token string) Result
}

type tokenVerifierService struct {
	repo repo.TokenRepository
}

func NewTokenVerifierService(r repo.TokenRepository) VerifierService {
	return &tokenVerifierService{repo: r}
}

func (s *tokenVerifierService) VerifyToken(token string) Result {
	if token == "" {
		return Result{Code: CodeBadRequest}
	}
	sub, err := s.repo.Verify(token)
	if err != nil {
		if errors.Is(err, repo.ErrInvalidToken) {
			return Result{Code: CodeInvalidToken}
		}
		return Result{Code: CodeInvalidToken}
	}
	return Result{Valid: true, Subject: sub, Code: CodeOK}
}
