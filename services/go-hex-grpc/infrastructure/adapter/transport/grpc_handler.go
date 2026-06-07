// Package grpc is the INCOMING adapter that drives the application
// over gRPC using the verifier.proto service definition.
package transport

import (
	"context"

	"github.com/dablon/arch-bench/proto/pb"
	"github.com/dablon/arch-bench/services/go-hex-grpc/domain/entity"
	"github.com/dablon/arch-bench/services/go-hex-grpc/domain/usecase"
)

type VerifyHandler struct {
	pb.UnimplementedVerifierServer
	uc *usecase.VerifyToken
}

func NewVerifyHandler(uc *usecase.VerifyToken) *VerifyHandler {
	return &VerifyHandler{uc: uc}
}

func (h *VerifyHandler) Verify(_ context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	res := h.uc.Execute(req.Token)
	return &pb.VerifyResponse{Valid: res.Valid, Subject: res.Subject, Code: string(res.Code)}, nil
}

var _ = entity.CodeOK // keep import
