package jwt

import (
	"errors"
	"testing"
	"time"

	gjwt "github.com/golang-jwt/jwt/v5"
)

const secret = "test"

func mintToken(t *testing.T, sub string, exp time.Duration) string {
	t.Helper()
	tok := gjwt.NewWithClaims(gjwt.SigningMethodHS256, gjwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(exp).Unix(),
	})
	s, _ := tok.SignedString([]byte(secret))
	return s
}

func TestAdapter_OK(t *testing.T) {
	a := NewHmacJwtAdapter(secret)
	sub, err := a.Verify(mintToken(t, "alice", time.Hour))
	if err != nil || sub != "alice" {
		t.Fatalf("err=%v sub=%q", err, sub)
	}
}

func TestAdapter_Bad(t *testing.T) {
	a := NewHmacJwtAdapter(secret)
	_, err := a.Verify("x")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestAdapter_BadAlg(t *testing.T) {
	a := NewHmacJwtAdapter(secret)
	_, err := a.Verify("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhbGljZSJ9.")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestAdapter_BadSecret(t *testing.T) {
	tok := gjwt.NewWithClaims(gjwt.SigningMethodHS256, gjwt.MapClaims{"sub": "x"})
	s, _ := tok.SignedString([]byte("wrong"))
	a := NewHmacJwtAdapter(secret)
	_, err := a.Verify(s)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestAdapter_Expired(t *testing.T) {
	a := NewHmacJwtAdapter(secret)
	_, err := a.Verify(mintToken(t, "alice", -time.Hour))
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}
