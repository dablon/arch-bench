// Package app is the application layer.
package app

import (
	"github.com/dablon/arch-bench/services/go-hex-grpc/domain/port"
	"github.com/dablon/arch-bench/services/go-hex-grpc/domain/usecase"
)

type Application struct {
	VerifyToken *usecase.VerifyToken
}

func NewApplication(verifier port.TokenVerifier) *Application {
	return &Application{
		VerifyToken: usecase.NewVerifyToken(verifier),
	}
}
