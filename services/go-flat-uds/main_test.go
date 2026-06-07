package main

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test"

func mintToken(t *testing.T, sub string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	s, _ := tok.SignedString([]byte(testSecret))
	return s
}

func startServer(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	sock := filepath.Join(dir, "t.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go handleConn(c, []byte(testSecret))
		}
	}()
	t.Cleanup(func() { l.Close(); os.Remove(sock) })
	return sock
}

func dial(t *testing.T, sock string) (net.Conn, *bufio.Reader) {
	t.Helper()
	c, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	return c, bufio.NewReader(c)
}

func TestVerifyOK(t *testing.T) {
	sock := startServer(t)
	c, r := dial(t, sock)
	defer c.Close()
	c.Write([]byte("VERIFY " + mintToken(t, "alice") + "\n"))
	line, _ := r.ReadString('\n')
	if !strings.HasPrefix(line, "OK alice") {
		t.Fatalf("got %q", line)
	}
}

func TestVerifyBadToken(t *testing.T) {
	sock := startServer(t)
	c, r := dial(t, sock)
	defer c.Close()
	c.Write([]byte("VERIFY garbage\n"))
	line, _ := r.ReadString('\n')
	if !strings.HasPrefix(line, "ERR INVALID_TOKEN") {
		t.Fatalf("got %q", line)
	}
}

func TestVerifyBadAlg(t *testing.T) {
	sock := startServer(t)
	c, r := dial(t, sock)
	defer c.Close()
	tok := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhbGljZSJ9."
	c.Write([]byte("VERIFY " + tok + "\n"))
	line, _ := r.ReadString('\n')
	if !strings.HasPrefix(line, "ERR") {
		t.Fatalf("got %q", line)
	}
}

func TestVerifyBadRequest(t *testing.T) {
	sock := startServer(t)
	c, r := dial(t, sock)
	defer c.Close()
	c.Write([]byte("HELLO\n"))
	line, _ := r.ReadString('\n')
	if !strings.HasPrefix(line, "ERR BAD_REQUEST") {
		t.Fatalf("got %q", line)
	}
}

func TestVerifyEmptyToken(t *testing.T) {
	sock := startServer(t)
	c, r := dial(t, sock)
	defer c.Close()
	c.Write([]byte("VERIFY \n"))
	line, _ := r.ReadString('\n')
	if !strings.HasPrefix(line, "ERR BAD_REQUEST") {
		t.Fatalf("got %q", line)
	}
}

func TestVerifyBadSecret(t *testing.T) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "x"})
	s, _ := tok.SignedString([]byte("wrong"))
	sock := startServer(t)
	c, r := dial(t, sock)
	defer c.Close()
	c.Write([]byte("VERIFY " + s + "\n"))
	line, _ := r.ReadString('\n')
	if !strings.HasPrefix(line, "ERR INVALID_TOKEN") {
		t.Fatalf("got %q", line)
	}
}
