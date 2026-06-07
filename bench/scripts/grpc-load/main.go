// gRPC load generator. Connects, sends Verify(token), records latency.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/dablon/arch-bench/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:50051", "gRPC address")
	token := flag.String("token", "", "JWT token")
	duration := flag.Int("duration", 20, "duration in seconds")
	flag.Parse()

	if *token == "" {
		log.Fatal("token required")
	}

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	stub := pb.NewVerifierClient(conn)

	end := time.Now().Add(time.Duration(*duration) * time.Second)
	total := 0
	for time.Now().Before(end) {
		t0 := time.Now()
		_, err := stub.Verify(context.Background(), &pb.VerifyRequest{Token: *token})
		t1 := time.Now()
		if err != nil {
			fmt.Fprintf(os.Stderr, "err: %v\n", err)
			continue
		}
		lat := float64(t1.Sub(t0)) / float64(time.Millisecond)
		fmt.Printf("%.3f\n", lat)
		total++
	}
	fmt.Fprintf(os.Stderr, "# grpc-go count=%d\n", total)
}
