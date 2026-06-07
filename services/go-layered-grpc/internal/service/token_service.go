// Package service is the business-logic layer for the layered gRPC verifier.
package service

import "github.com/dablon/arch-bench/services/go-layered-grpc/internal/repo"

type ResultCode string

const (
	CodeOK           ResultCode = "OK"
	CodeBadRequest   ResultCode = "ERR_BAD_REQUEST"
	CodeInvalidToken ResultCode = "ERR_INVALID_TOKEN"
)

type Result struct {
	Valid   bool
	Subject string
	Code    ResultCode
}

type VerifierService interface {
	VerifyToken(token string) Result
}

type tokenVerifierService struct{ repo repo.TokenRepository }

func NewTokenVerifierService(r repo.TokenRepository) VerifierService {
	return &tokenVerifierService{repo: r}
}

func (s *tokenVerifierService) VerifyToken(token string) Result {
	if token == "" {
		return Result{Code: CodeBadRequest}
	}
	sub, err := s.repo.Verify(token)
	if err != nil {
		return Result{Code: CodeInvalidToken}
	}
	return Result{Valid: true, Subject: sub, Code: CodeOK}
}
