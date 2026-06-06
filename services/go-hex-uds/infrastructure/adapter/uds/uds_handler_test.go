package uds

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dablon/arch-bench/services/go-hex-uds/domain/entity"
	"github.com/dablon/arch-bench/services/go-hex-uds/domain/port"
	"github.com/dablon/arch-bench/services/go-hex-uds/domain/usecase"
)

type fakePort struct{ sub string; err error }

func (f *fakePort) Verify(_ string) (string, error) { return f.sub, f.err }

func start(t *testing.T, p port.TokenVerifier) string {
	t.Helper()
	dir := t.TempDir()
	sock := filepath.Join(dir, "t.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	h := NewHandler(usecase.NewVerifyToken(p))
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go h.Handle(c)
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

func TestOK(t *testing.T) {
	sock := start(t, &fakePort{sub: "alice"})
	c, r := dial(t, sock)
	defer c.Close()
	c.Write([]byte("VERIFY x\n"))
	line, _ := r.ReadString('\n')
	if !strings.HasPrefix(line, "OK alice") {
		t.Fatalf("got %q", line)
	}
}

func TestInvalid(t *testing.T) {
	sock := start(t, &fakePort{err: errAny()})
	c, r := dial(t, sock)
	defer c.Close()
	c.Write([]byte("VERIFY x\n"))
	line, _ := r.ReadString('\n')
	if !strings.HasPrefix(line, "ERR INVALID_TOKEN") {
		t.Fatalf("got %q", line)
	}
}

func TestEmpty(t *testing.T) {
	sock := start(t, &fakePort{sub: "alice"})
	c, r := dial(t, sock)
	defer c.Close()
	c.Write([]byte("VERIFY \n"))
	line, _ := r.ReadString('\n')
	if !strings.HasPrefix(line, "ERR BAD_REQUEST") {
		t.Fatalf("got %q", line)
	}
}

func TestBadLine(t *testing.T) {
	sock := start(t, &fakePort{sub: "alice"})
	c, r := dial(t, sock)
	defer c.Close()
	c.Write([]byte("HELLO\n"))
	line, _ := r.ReadString('\n')
	if !strings.HasPrefix(line, "ERR BAD_REQUEST") {
		t.Fatalf("got %q", line)
	}
}

func errAny() error { return fakeErr("x") }

type fakeErr string

func (e fakeErr) Error() string { return string(e) }

var _ = entity.CodeOK
