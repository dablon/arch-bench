package transport

import (
	"context"
	"errors"
	"testing"

	"github.com/dablon/arch-bench/proto/pb"
	"github.com/dablon/arch-bench/services/go-hex-grpc/domain/port"
	"github.com/dablon/arch-bench/services/go-hex-grpc/domain/usecase"
)

type fakePort struct{ sub string; err error }

func (f *fakePort) Verify(_ string) (string, error) { return f.sub, f.err }

func TestVerifyOK(t *testing.T) {
	h := NewVerifyHandler(usecase.NewVerifyToken(&fakePort{sub: "alice"}))
	r, err := h.Verify(context.Background(), &pb.VerifyRequest{Token: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if !r.Valid || r.Subject != "alice" || r.Code != "OK" {
		t.Fatalf("got %+v", r)
	}
}

func TestVerifyInvalid(t *testing.T) {
	h := NewVerifyHandler(usecase.NewVerifyToken(&fakePort{err: errors.New("nope")}))
	r, _ := h.Verify(context.Background(), &pb.VerifyRequest{Token: "x"})
	if r.Code != "ERR_INVALID_TOKEN" {
		t.Fatalf("got %+v", r)
	}
}

func TestVerifyEmpty(t *testing.T) {
	h := NewVerifyHandler(usecase.NewVerifyToken(&fakePort{sub: "alice"}))
	r, _ := h.Verify(context.Background(), &pb.VerifyRequest{Token: ""})
	if r.Code != "ERR_BAD_REQUEST" {
		t.Fatalf("got %+v", r)
	}
}

// _ keeps port imported.
var _ port.TokenVerifier
