// Command server is the composition root for go-hex-uds.
package main

import (
	"log"
	"net"
	"os"
	"sync"

	"github.com/dablon/arch-bench/services/go-hex-uds/app"
	"github.com/dablon/arch-bench/services/go-hex-uds/infrastructure/adapter/jwt"
	"github.com/dablon/arch-bench/services/go-hex-uds/infrastructure/adapter/uds"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	path := os.Getenv("UDS_PATH")
	if path == "" {
		path = "/tmp/go-hex-uds.sock"
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
	log.Printf("go-hex-uds listening on %s", path)

	// Composition: wire adapter → port, then use case, then handler.
	jwtAdapter := jwt.NewHmacJwtAdapter(secret)
	application := app.NewApplication(jwtAdapter)
	handler := uds.NewHandler(application.VerifyToken)

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
			handler.Handle(c)
		}()
	}
}
