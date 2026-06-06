package transport

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/dablon/arch-bench/services/go-hex-http/app"
	"github.com/dablon/arch-bench/services/go-hex-http/domain/entity"
)

type fakePort struct {
	sub string
	err error
}

func (f *fakePort) Verify(_ string) (string, error) { return f.sub, f.err }

func newApp() *app.Application {
	return app.NewApplication(&fakePort{sub: "alice"})
}

func TestHealth(t *testing.T) {
	w := httptest.NewRecorder()
	HealthHandler(w, httptest.NewRequest("GET", "/health", nil))
	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyOK(t *testing.T) {
	a := app.NewApplication(&fakePort{sub: "alice"})
	h := NewHandler(a)
	body, _ := json.Marshal(verifyRequest{Token: "x"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestVerifyBadMethod(t *testing.T) {
	h := NewHandler(newApp())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/verify", nil))
	if w.Code != 405 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyBadBody(t *testing.T) {
	h := NewHandler(newApp())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader([]byte("nope"))))
	if w.Code != 400 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyEmpty(t *testing.T) {
	a := app.NewApplication(&fakePort{sub: "alice"})
	h := NewHandler(a)
	body, _ := json.Marshal(verifyRequest{Token: ""})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyInvalid(t *testing.T) {
	a := app.NewApplication(&fakePort{err: errAny()})
	h := NewHandler(a)
	body, _ := json.Marshal(verifyRequest{Token: "x"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader(body)))
	if w.Code != 401 {
		t.Fatalf("status=%d", w.Code)
	}
}

// errAny is a tiny helper so we don't need a new import in the test file.
func errAny() error {
	return fakeError("boom")
}

type fakeError string

func (e fakeError) Error() string { return string(e) }

// _ keeps entity imported for clarity.
var _ = entity.CodeOK
