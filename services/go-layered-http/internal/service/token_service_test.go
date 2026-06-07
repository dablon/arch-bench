package service

import (
	"errors"
	"testing"

	"github.com/dablon/arch-bench/services/go-layered-http/internal/repo"
)

// fakeRepo lets us test the service without a real JWT library.
type fakeRepo struct {
	sub  string
	err  error
}

func (f *fakeRepo) Verify(_ string) (string, error) {
	return f.sub, f.err
}

func TestService_Empty(t *testing.T) {
	s := NewTokenVerifierService(&fakeRepo{})
	r := s.VerifyToken("")
	if r.Code != CodeBadRequest {
		t.Fatalf("got %+v", r)
	}
}

func TestService_InvalidFromRepo(t *testing.T) {
	s := NewTokenVerifierService(&fakeRepo{err: repo.ErrInvalidToken})
	r := s.VerifyToken("x")
	if r.Valid {
		t.Fatalf("got %+v", r)
	}
	if r.Code != CodeInvalidToken {
		t.Fatalf("got %+v", r)
	}
}

func TestService_GenericErrorFromRepo(t *testing.T) {
	s := NewTokenVerifierService(&fakeRepo{err: errors.New("boom")})
	r := s.VerifyToken("x")
	if r.Code != CodeInvalidToken {
		t.Fatalf("got %+v", r)
	}
}

func TestService_OK(t *testing.T) {
	s := NewTokenVerifierService(&fakeRepo{sub: "alice"})
	r := s.VerifyToken("x")
	if !r.Valid || r.Subject != "alice" || r.Code != CodeOK {
		t.Fatalf("got %+v", r)
	}
}
