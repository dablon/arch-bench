package handler

import (
	"context"
	"testing"

	"github.com/dablon/arch-bench/proto/pb"
	"github.com/dablon/arch-bench/services/go-layered-grpc/internal/service"
)

type stubSvc struct{ r service.Result }

func (s *stubSvc) VerifyToken(_ string) service.Result { return s.r }

func TestVerifyOK(t *testing.T) {
	h := NewVerifyHandler(&stubSvc{r: service.Result{Valid: true, Subject: "alice", Code: service.CodeOK}})
	r, err := h.Verify(context.Background(), &pb.VerifyRequest{Token: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if !r.Valid || r.Subject != "alice" || r.Code != "OK" {
		t.Fatalf("%+v", r)
	}
}

func TestVerifyInvalid(t *testing.T) {
	h := NewVerifyHandler(&stubSvc{r: service.Result{Code: service.CodeInvalidToken}})
	r, _ := h.Verify(context.Background(), &pb.VerifyRequest{Token: "x"})
	if r.Code != "ERR_INVALID_TOKEN" {
		t.Fatalf("%+v", r)
	}
}

func TestVerifyBadRequest(t *testing.T) {
	h := NewVerifyHandler(&stubSvc{r: service.Result{Code: service.CodeBadRequest}})
	r, _ := h.Verify(context.Background(), &pb.VerifyRequest{Token: ""})
	if r.Code != "ERR_BAD_REQUEST" {
		t.Fatalf("%+v", r)
	}
}
