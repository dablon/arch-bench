// Package port holds the interfaces the application needs.
package port

type TokenVerifier interface {
	Verify(token string) (subject string, err error)
}
