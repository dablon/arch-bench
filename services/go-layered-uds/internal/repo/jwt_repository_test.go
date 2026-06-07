package repo

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secret = "test"

func mintToken(t *testing.T, sub string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	s, _ := tok.SignedString([]byte(secret))
	return s
}

func TestRepo_OK(t *testing.T) {
	r := NewJwtRepository(secret)
	sub, err := r.Verify(mintToken(t, "alice"))
	if err != nil || sub != "alice" {
		t.Fatalf("got sub=%q err=%v", sub, err)
	}
}

func TestRepo_Bad(t *testing.T) {
	r := NewJwtRepository(secret)
	_, err := r.Verify("garbage")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("err=%v", err)
	}
}

func TestRepo_BadAlg(t *testing.T) {
	r := NewJwtRepository(secret)
	tok := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhbGljZSJ9."
	_, err := r.Verify(tok)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("err=%v", err)
	}
}

func TestRepo_BadSecret(t *testing.T) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "x"})
	s, _ := tok.SignedString([]byte("wrong"))
	r := NewJwtRepository(secret)
	_, err := r.Verify(s)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("err=%v", err)
	}
}
