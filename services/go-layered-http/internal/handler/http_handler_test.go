package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/dablon/arch-bench/services/go-layered-http/internal/service"
)

type stubSvc struct {
	r service.Result
}

func (s *stubSvc) VerifyToken(_ string) service.Result { return s.r }

func newHandler(r service.Result) *VerifyHandler {
	return NewVerifyHandler(&stubSvc{r: r})
}

func TestHealth(t *testing.T) {
	w := httptest.NewRecorder()
	HealthHandler(w, httptest.NewRequest("GET", "/health", nil))
	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyOK(t *testing.T) {
	h := newHandler(service.Result{Valid: true, Subject: "alice", Code: service.CodeOK})
	body, _ := json.Marshal(verifyRequest{Token: "x"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestVerifyBadMethod(t *testing.T) {
	h := newHandler(service.Result{})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/verify", nil))
	if w.Code != 405 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyBadBody(t *testing.T) {
	h := newHandler(service.Result{})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader([]byte("nope"))))
	if w.Code != 400 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyBadRequestFromService(t *testing.T) {
	h := newHandler(service.Result{Code: service.CodeBadRequest})
	body, _ := json.Marshal(verifyRequest{Token: ""})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyInvalidFromService(t *testing.T) {
	h := newHandler(service.Result{Code: service.CodeInvalidToken})
	body, _ := json.Marshal(verifyRequest{Token: "x"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader(body)))
	if w.Code != 401 {
		t.Fatalf("status=%d", w.Code)
	}
}
