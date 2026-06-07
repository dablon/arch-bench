// Package jwt is the OUTGOING adapter that implements port.TokenVerifier
// using HMAC-SHA256 JWT verification.
//
// An "outgoing adapter" in hexagonal terms: it's the bridge from the
// application's port to an external capability. Here, the external
// capability is "given a string, tell me if it's a valid JWT signed
// with my secret, and what subject is on it".
//
// The implementation uses github.com/golang-jwt, but the rest of the
// codebase does not import this package — they import the port.
package jwt

import (
	"errors"
	"fmt"

	gjwt "github.com/golang-jwt/jwt/v5"
)

// ErrInvalid is what the adapter returns on any verification failure.
// The use case treats all errors as "invalid"; the adapter collapses
// the many failure modes (bad alg, bad sig, expired) into one.
var ErrInvalid = errors.New("invalid token")

type HmacJwtAdapter struct {
	secret []byte
}

func NewHmacJwtAdapter(secret string) *HmacJwtAdapter {
	return &HmacJwtAdapter{secret: []byte(secret)}
}

func (a *HmacJwtAdapter) Verify(token string) (string, error) {
	parsed, err := gjwt.Parse(token, func(t *gjwt.Token) (any, error) {
		if _, ok := t.Method.(*gjwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("bad alg")
		}
		return a.secret, nil
	})
	if err != nil || !parsed.Valid {
		return "", ErrInvalid
	}
	claims, _ := parsed.Claims.(gjwt.MapClaims)
	sub, _ := claims["sub"].(string)
	return sub, nil
}
