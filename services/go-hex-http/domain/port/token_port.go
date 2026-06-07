// Package port contains the interfaces the application needs to talk
// to the outside world.
//
// In hexagonal architecture, "ports" are the abstract dependencies the
// domain has on infrastructure. Adapters (in infrastructure/) implement
// these ports. The domain doesn't know who the adapters are.
package port

// TokenVerifier is the OUTGOING port: the application needs a way to
// verify a token. It does not care whether that verification is done
// with HMAC-SHA256, PASETO, a remote service, or a magic 8-ball.
//
// Returning a structured error rather than a Result here is the
// hexagonal convention: ports express the contract in terms of the
// caller-friendly outcome. The use case translates that into a
// domain Result.
type TokenVerifier interface {
	Verify(token string) (subject string, err error)
}
