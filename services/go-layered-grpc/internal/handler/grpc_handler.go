// Package handler is the gRPC-transport layer for the layered
// architecture. It knows the proto types and the service interface
// and nothing else.
package handler

import (
	"context"

	"github.com/dablon/arch-bench/proto/pb"
	"github.com/dablon/arch-bench/services/go-layered-grpc/internal/service"
)

type VerifyHandler struct {
	pb.UnimplementedVerifierServer
	svc service.VerifierService
}

func NewVerifyHandler(s service.VerifierService) *VerifyHandler {
	return &VerifyHandler{svc: s}
}

func (h *VerifyHandler) Verify(_ context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	res := h.svc.VerifyToken(req.Token)
	return &pb.VerifyResponse{Valid: res.Valid, Subject: res.Subject, Code: string(res.Code)}, nil
}
