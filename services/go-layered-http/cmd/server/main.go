// Command server wires the layered architecture for go-layered-http.
//
// Architectural style: LAYERED (a.k.a. "n-tier", "3-tier").
//
// Layers (top-down call direction, bottom-up dependency):
//   1. handler/  — HTTP-specific glue (parse JSON, write response, status codes)
//   2. service/  — business logic (verify a token, return a domain result)
//   3. repo/     — data access (the JWT verifier adapter against the jwt library)
//
// Each layer is a package. Imports go strictly downward:
//   handler imports service
//   service imports repo
//   repo imports nothing internal
//
// The service has a TokenRepository interface (defined in repo/); the
// handler has a VerifierService interface (defined in service/). This
// means: a future swap of "JWT" for "PASETO" or "Macaroon" only touches
// repo/. A future swap of "HTTP" for "gRPC" only touches handler/.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dablon/arch-bench/services/go-layered-http/internal/handler"
	"github.com/dablon/arch-bench/services/go-layered-http/internal/repo"
	"github.com/dablon/arch-bench/services/go-layered-http/internal/service"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8081"
	}
	if secret == "" {
		log.Fatal("JWT_SECRET required")
	}

	// Composition root: wire layers bottom-up.
	tokenRepo := repo.NewJwtRepository(secret)
	verifierSvc := service.NewTokenVerifierService(tokenRepo)
	httpHandler := handler.NewVerifyHandler(verifierSvc)

	mux := http.NewServeMux()
	mux.Handle("/verify", httpHandler)
	mux.HandleFunc("/health", handler.HealthHandler)

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("go-layered-http listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}
