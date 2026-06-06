// Package repo is the data-access layer of the layered architecture.
//
// In a layered system the "repository" abstracts a data source. In a
// stateless verifier, the "data" is a JWT. The repository hides how
// tokens are decoded and what crypto library is used.
package repo

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is the sentinel error returned when a token fails
// any verification step. Service-layer code can check for this with
// errors.Is.
var ErrInvalidToken = errors.New("invalid token")

// TokenRepository is the port the service layer depends on. Defined
// here (in the layer that implements it) so the service package can
// import it without an import cycle.
type TokenRepository interface {
	// Verify parses and validates a token, returning the subject on
	// success. Returns ErrInvalidToken on any failure.
	Verify(token string) (subject string, err error)
}

// JwtRepository is a TokenRepository backed by github.com/golang-jwt.
type JwtRepository struct {
	secret []byte
}

func NewJwtRepository(secret string) *JwtRepository {
	return &JwtRepository{secret: []byte(secret)}
}

func (r *JwtRepository) Verify(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("%w: empty", ErrInvalidToken)
	}
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("bad alg")
		}
		return r.secret, nil
	})
	if err != nil || !parsed.Valid {
		return "", ErrInvalidToken
	}
	claims, _ := parsed.Claims.(jwt.MapClaims)
	sub, _ := claims["sub"].(string)
	return sub, nil
}
