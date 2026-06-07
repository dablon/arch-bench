package usecase

import (
	"errors"
	"testing"

	"github.com/dablon/arch-bench/services/go-hex-grpc/domain/entity"
)

type fakePort struct{ sub string; err error }

func (f *fakePort) Verify(_ string) (string, error) { return f.sub, f.err }

func TestExecute_Empty(t *testing.T) {
	u := NewVerifyToken(&fakePort{})
	r := u.Execute("")
	if r.Code != entity.CodeBadRequest {
		t.Fatalf("got %+v", r)
	}
}

func TestExecute_Invalid(t *testing.T) {
	u := NewVerifyToken(&fakePort{err: errors.New("nope")})
	r := u.Execute("x")
	if r.Code != entity.CodeInvalidToken {
		t.Fatalf("got %+v", r)
	}
}

func TestExecute_OK(t *testing.T) {
	u := NewVerifyToken(&fakePort{sub: "alice"})
	r := u.Execute("x")
	if !r.Valid || r.Subject != "alice" || r.Code != entity.CodeOK {
		t.Fatalf("got %+v", r)
	}
}
