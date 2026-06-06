// Package port holds the interfaces the application needs.
package port

// TokenVerifier is the OUTGOING port: the application needs a way to
// verify a token. The implementation is supplied by an adapter in
// infrastructure/.
type TokenVerifier interface {
	Verify(token string) (subject string, err error)
}
