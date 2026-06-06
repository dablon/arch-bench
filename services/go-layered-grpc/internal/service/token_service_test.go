package service

import (
	"errors"
	"testing"

	"github.com/dablon/arch-bench/services/go-layered-grpc/internal/repo"
)

type fakeRepo struct{ sub string; err error }

func (f *fakeRepo) Verify(_ string) (string, error) { return f.sub, f.err }

func TestService_Empty(t *testing.T) {
	s := NewTokenVerifierService(&fakeRepo{})
	r := s.VerifyToken("")
	if r.Code != CodeBadRequest {
		t.Fatalf("%+v", r)
	}
}

func TestService_Invalid(t *testing.T) {
	s := NewTokenVerifierService(&fakeRepo{err: repo.ErrInvalidToken})
	r := s.VerifyToken("x")
	if r.Code != CodeInvalidToken {
		t.Fatalf("%+v", r)
	}
}

func TestService_GenericError(t *testing.T) {
	s := NewTokenVerifierService(&fakeRepo{err: errors.New("nope")})
	r := s.VerifyToken("x")
	if r.Code != CodeInvalidToken {
		t.Fatalf("%+v", r)
	}
}

func TestService_OK(t *testing.T) {
	s := NewTokenVerifierService(&fakeRepo{sub: "alice"})
	r := s.VerifyToken("x")
	if !r.Valid || r.Subject != "alice" || r.Code != CodeOK {
		t.Fatalf("%+v", r)
	}
}
