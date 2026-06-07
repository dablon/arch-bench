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

func TestVerify_OK(t *testing.T) {
	a := NewHmacJwtAdapter(secret)
	sub, err := a.Verify(mintToken(t, "alice", time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if sub != "alice" {
		t.Fatalf("sub=%q", sub)
	}
}

func TestVerify_BadToken(t *testing.T) {
	a := NewHmacJwtAdapter(secret)
	_, err := a.Verify("garbage")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestVerify_BadAlg(t *testing.T) {
	a := NewHmacJwtAdapter(secret)
	tok := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhbGljZSJ9."
	_, err := a.Verify(tok)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestVerify_BadSecret(t *testing.T) {
	tok := gjwt.NewWithClaims(gjwt.SigningMethodHS256, gjwt.MapClaims{"sub": "x"})
	s, _ := tok.SignedString([]byte("wrong"))
	a := NewHmacJwtAdapter(secret)
	_, err := a.Verify(s)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}

func TestVerify_Expired(t *testing.T) {
	a := NewHmacJwtAdapter(secret)
	_, err := a.Verify(mintToken(t, "alice", -time.Hour))
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v", err)
	}
}
