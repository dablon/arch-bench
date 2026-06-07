package app

import (
	"errors"
	"testing"

	"github.com/dablon/arch-bench/services/go-hex-http/domain/entity"
)

type fakePort struct{ err error }

func (f *fakePort) Verify(_ string) (string, error) { return "", f.err }

func TestApplication_PassesPortToUseCase(t *testing.T) {
	a := NewApplication(&fakePort{err: errors.New("nope")})
	r := a.VerifyToken.Execute("x")
	if r.Code != entity.CodeInvalidToken {
		t.Fatalf("got %+v", r)
	}
}
