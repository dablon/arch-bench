// Command server wires the layered UDS verifier.
package main

import (
	"log"
	"net"
	"os"
	"sync"

	"github.com/dablon/arch-bench/services/go-layered-uds/internal/handler"
	"github.com/dablon/arch-bench/services/go-layered-uds/internal/repo"
	"github.com/dablon/arch-bench/services/go-layered-uds/internal/service"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	path := os.Getenv("UDS_PATH")
	if path == "" {
		path = "/tmp/go-layered-uds.sock"
	}
	if secret == "" {
		log.Fatal("JWT_SECRET required")
	}
	_ = os.Remove(path)
	l, err := net.Listen("unix", path)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	_ = os.Chmod(path, 0660)
	log.Printf("go-layered-uds listening on %s", path)

	tokenRepo := repo.NewJwtRepository(secret)
	verifierSvc := service.NewTokenVerifierService(tokenRepo)
	h := handler.NewHandler(verifierSvc)

	var wg sync.WaitGroup
	for {
		c, err := l.Accept()
		if err != nil {
			wg.Wait()
			return
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.Handle(c)
		}()
	}
}
