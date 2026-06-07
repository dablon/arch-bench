// Package jwt is the OUTGOING adapter that implements the TokenVerifier
// port using HMAC-SHA256 JWT.
package jwt

import (
	"errors"
	"fmt"

	gjwt "github.com/golang-jwt/jwt/v5"
)

var ErrInvalid = errors.New("invalid token")

type HmacJwtAdapter struct{ secret []byte }

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
