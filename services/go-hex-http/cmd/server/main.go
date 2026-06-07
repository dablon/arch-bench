// Command server is the composition root for go-hex-http.
//
// Architectural style: HEXAGONAL (a.k.a. Ports & Adapters,
// "Clean Architecture" in the Uncle Bob sense).
//
// The code is partitioned by *direction of dependency*, not by layer
// type. The dependency rule:
//
//   domain/            — knows nothing about anything. Pure Go.
//     entity/             — Value objects: Token, VerificationResult
//     port/               — Interfaces the application needs
//     usecase/            — Application-specific business rules
//   infrastructure/    — knows about domain. Adapters live here.
//     adapter/http/       — incoming adapter (HTTP in, calls the use case)
//     adapter/jwt/        — outgoing adapter (HMAC JWT crypto)
//   app/               — composition root: wires adapters to ports
//
// Arrows point inward, always:
//   infrastructure/ ──> domain/
//   app/           ──> domain/  (and the concrete adapters it composes)
//
// Swap "jwt" for "paseto" in adapter/. Swap "http" for "grpc" in
// adapter/. The use case and entities do not change.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dablon/arch-bench/services/go-hex-http/app"
	jwtadapter "github.com/dablon/arch-bench/services/go-hex-http/infrastructure/adapter/jwt"
	transport "github.com/dablon/arch-bench/services/go-hex-http/infrastructure/adapter/transport"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8082"
	}
	if secret == "" {
		log.Fatal("JWT_SECRET required")
	}

	// Build the application (a use case + a port the use case depends on).
	jwtAdapter := jwtadapter.NewHmacJwtAdapter(secret)
	application := app.NewApplication(jwtAdapter)

	// Build the primary adapter (HTTP) that drives the use case.
	handler := transport.NewHandler(application)
	mux := http.NewServeMux()
	mux.Handle("/verify", handler)
	mux.HandleFunc("/health", transport.HealthHandler)

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("go-hex-http listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}
