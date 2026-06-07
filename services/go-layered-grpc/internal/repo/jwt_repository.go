// Package repo is the data-access layer for the layered gRPC verifier.
package repo

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type TokenRepository interface {
	Verify(token string) (subject string, err error)
}

type JwtRepository struct{ secret []byte }

func NewJwtRepository(secret string) *JwtRepository {
	return &JwtRepository{secret: []byte(secret)}
}

func (r *JwtRepository) Verify(token string) (string, error) {
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
