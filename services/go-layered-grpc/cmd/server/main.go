// Command server is the composition root for the layered gRPC verifier.
package main

import (
	"log"
	"net"
	"os"

	"github.com/dablon/arch-bench/proto/pb"
	"github.com/dablon/arch-bench/services/go-layered-grpc/internal/handler"
	"github.com/dablon/arch-bench/services/go-layered-grpc/internal/repo"
	"github.com/dablon/arch-bench/services/go-layered-grpc/internal/service"
	"google.golang.org/grpc"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = "127.0.0.1:50052"
	}
	if secret == "" {
		log.Fatal("JWT_SECRET required")
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	tokenRepo := repo.NewJwtRepository(secret)
	svc := service.NewTokenVerifierService(tokenRepo)
	h := handler.NewVerifyHandler(svc)

	s := grpc.NewServer()
	pb.RegisterVerifierServer(s, h)
	log.Printf("go-layered-grpc listening on %s", addr)
	log.Fatal(s.Serve(l))
}
