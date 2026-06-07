// Command server is the composition root for go-hex-grpc.
package main

import (
	"log"
	"net"
	"os"

	"github.com/dablon/arch-bench/proto/pb"
	"github.com/dablon/arch-bench/services/go-hex-grpc/app"
	"github.com/dablon/arch-bench/services/go-hex-grpc/infrastructure/adapter/transport"
	"github.com/dablon/arch-bench/services/go-hex-grpc/infrastructure/adapter/jwt"
	"google.golang.org/grpc"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = "127.0.0.1:50053"
	}
	if secret == "" {
		log.Fatal("JWT_SECRET required")
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	jwtAdapter := jwt.NewHmacJwtAdapter(secret)
	application := app.NewApplication(jwtAdapter)
	handler := transport.NewVerifyHandler(application.VerifyToken)

	s := grpc.NewServer()  // stdlib name shadow; careful in this package
	pb.RegisterVerifierServer(s, handler)
	log.Printf("go-hex-grpc listening on %s", addr)
	log.Fatal(s.Serve(l))
}
