package usecase

import (
	"errors"
	"testing"

	"github.com/dablon/arch-bench/services/go-hex-http/domain/entity"
)

type fakeVerifier struct {
	sub string
	err error
}

func (f *fakeVerifier) Verify(_ string) (string, error) {
	return f.sub, f.err
}

func TestExecute_Empty(t *testing.T) {
	u := NewVerifyToken(&fakeVerifier{})
	r := u.Execute("")
	if r.Code != entity.CodeBadRequest {
		t.Fatalf("got %+v", r)
	}
}

func TestExecute_InvalidFromPort(t *testing.T) {
	u := NewVerifyToken(&fakeVerifier{err: errors.New("boom")})
	r := u.Execute("x")
	if r.Code != entity.CodeInvalidToken {
		t.Fatalf("got %+v", r)
	}
}

func TestExecute_OK(t *testing.T) {
	u := NewVerifyToken(&fakeVerifier{sub: "alice"})
	r := u.Execute("x")
	if !r.Valid || r.Subject != "alice" || r.Code != entity.CodeOK {
		t.Fatalf("got %+v", r)
	}
}
