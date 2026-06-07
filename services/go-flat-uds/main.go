// Package main is a FLAT-architecture JWT verifier over Unix Domain Socket.
//
// Architectural style: FLAT. Single file, no abstractions, all logic
// inline. The transport (UDS line protocol) and the JWT logic share
// the same package, the same function, the same scope.
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

const (
	maxLine = 4096
)

func handleConn(c net.Conn, secret []byte) {
	defer c.Close()
	r := bufio.NewReader(c)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		var resp string
		if !strings.HasPrefix(line, "VERIFY ") {
			resp = "ERR BAD_REQUEST\n"
		} else {
			tok := strings.TrimPrefix(line, "VERIFY ")
			if tok == "" {
				resp = "ERR BAD_REQUEST\n"
				_, _ = c.Write([]byte(resp))
				continue
			}
			parsed, perr := jwt.Parse(tok, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("bad alg")
				}
				return secret, nil
			})
			if perr != nil || !parsed.Valid {
				resp = "ERR INVALID_TOKEN\n"
			} else {
				claims, _ := parsed.Claims.(jwt.MapClaims)
				sub, _ := claims["sub"].(string)
				resp = fmt.Sprintf("OK %s\n", sub)
			}
		}
		_, _ = c.Write([]byte(resp))
	}
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	path := os.Getenv("UDS_PATH")
	if path == "" {
		path = "/tmp/go-flat-uds.sock"
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
	log.Printf("go-flat-uds listening on %s", path)

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
			handleConn(c, []byte(secret))
		}()
	}
}
