package main

import (
	"context"
	"testing"
	"time"

	"github.com/dablon/arch-bench/proto/pb"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test"

func mintToken(t *testing.T, sub string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	s, _ := tok.SignedString([]byte(testSecret))
	return s
}

func newServer() *server { return &server{secret: []byte(testSecret)} }

func TestVerifyOK(t *testing.T) {
	s := newServer()
	r, err := s.Verify(context.Background(), &pb.VerifyRequest{Token: mintToken(t, "alice")})
	if err != nil {
		t.Fatal(err)
	}
	if !r.Valid || r.Code != "OK" || r.Subject != "alice" {
		t.Fatalf("got %+v", r)
	}
}

func TestVerifyEmpty(t *testing.T) {
	r, _ := newServer().Verify(context.Background(), &pb.VerifyRequest{Token: ""})
	if r.Code != "ERR_BAD_REQUEST" {
		t.Fatalf("%+v", r)
	}
}

func TestVerifyBad(t *testing.T) {
	r, _ := newServer().Verify(context.Background(), &pb.VerifyRequest{Token: "x"})
	if r.Code != "ERR_INVALID_TOKEN" {
		t.Fatalf("%+v", r)
	}
}

func TestVerifyBadAlg(t *testing.T) {
	r, _ := newServer().Verify(context.Background(), &pb.VerifyRequest{Token: "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhbGljZSJ9."})
	if r.Code != "ERR_INVALID_TOKEN" {
		t.Fatalf("%+v", r)
	}
}

func TestVerifyBadSecret(t *testing.T) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "x"})
	s, _ := tok.SignedString([]byte("wrong"))
	r, _ := newServer().Verify(context.Background(), &pb.VerifyRequest{Token: s})
	if r.Code != "ERR_INVALID_TOKEN" {
		t.Fatalf("%+v", r)
	}
}

func TestVerifyExpired(t *testing.T) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "x",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	s, _ := tok.SignedString([]byte(testSecret))
	r, _ := newServer().Verify(context.Background(), &pb.VerifyRequest{Token: s})
	if r.Code != "ERR_INVALID_TOKEN" {
		t.Fatalf("%+v", r)
	}
}
