// Package app is the application layer.
//
// In hexagonal/Clean terms, the "application" is the bundle of use
// cases the system exposes. This file is the composition root for
// the application object itself — it does not know which adapters
// are wired in; it only knows which ports it depends on.
package app

import (
	"github.com/dablon/arch-bench/services/go-hex-http/domain/port"
	"github.com/dablon/arch-bench/services/go-hex-http/domain/usecase"
)

// Application is the public surface of the system. A primary adapter
// (HTTP, gRPC, CLI) drives the use case through this struct.
type Application struct {
	VerifyToken *usecase.VerifyToken
}

func NewApplication(verifier port.TokenVerifier) *Application {
	return &Application{
		VerifyToken: usecase.NewVerifyToken(verifier),
	}
}
