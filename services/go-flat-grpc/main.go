// Package main is a FLAT-architecture JWT verifier over gRPC.
//
// Architectural style: FLAT. Single file, no abstractions, all logic
// inline. The proto service is implemented directly on a struct; the
// gRPC server, the JWT decoding, the proto types, the env config all
// share the same scope.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/dablon/arch-bench/proto/pb"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedVerifierServer
	secret []byte
}

// Verify is the single gRPC method. The whole pipeline lives here.
func (s *server) Verify(_ context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	if req.Token == "" {
		return &pb.VerifyResponse{Valid: false, Code: "ERR_BAD_REQUEST"}, nil
	}
	parsed, err := jwt.Parse(req.Token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("bad alg")
		}
		return s.secret, nil
	})
	if err != nil || !parsed.Valid {
		return &pb.VerifyResponse{Valid: false, Code: "ERR_INVALID_TOKEN"}, nil
	}
	claims, _ := parsed.Claims.(jwt.MapClaims)
	sub, _ := claims["sub"].(string)
	return &pb.VerifyResponse{Valid: true, Subject: sub, Code: "OK"}, nil
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = "127.0.0.1:50051"
	}
	if secret == "" {
		log.Fatal("JWT_SECRET required")
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	s := grpc.NewServer()
	pb.RegisterVerifierServer(s, &server{secret: []byte(secret)})
	log.Printf("go-flat-grpc listening on %s", addr)
	log.Fatal(s.Serve(l))
}
