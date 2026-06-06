package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test"

func mintToken(t *testing.T, sub string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  sub,
		"exp":  time.Now().Add(time.Hour).Unix(),
		"role": "user",
	})
	s, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func newMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		handleVerify(w, r, testSecret)
	})
	mux.HandleFunc("/health", handleHealth)
	return mux
}

func TestHealth(t *testing.T) {
	w := httptest.NewRecorder()
	newMux().ServeHTTP(w, httptest.NewRequest("GET", "/health", nil))
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestVerifyOK(t *testing.T) {
	body, _ := json.Marshal(verifyRequest{Token: mintToken(t, "alice")})
	w := httptest.NewRecorder()
	newMux().ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var out verifyResponse
	_ = json.NewDecoder(w.Body).Decode(&out)
	if !out.Valid || out.Code != "OK" || out.Subject != "alice" {
		t.Fatalf("got %+v", out)
	}
}

func TestVerifyBadToken(t *testing.T) {
	body, _ := json.Marshal(verifyRequest{Token: "***"})
	w := httptest.NewRecorder()
	newMux().ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader(body)))
	if w.Code != 401 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyBadMethod(t *testing.T) {
	w := httptest.NewRecorder()
	newMux().ServeHTTP(w, httptest.NewRequest("GET", "/verify", nil))
	if w.Code != 405 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyBadBody(t *testing.T) {
	w := httptest.NewRecorder()
	newMux().ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader([]byte("not json"))))
	if w.Code != 400 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyBadAlg(t *testing.T) {
	tok := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhbGljZSIsImV4cCI6OTk5OTk5OTk5OX0."
	body, _ := json.Marshal(verifyRequest{Token: tok})
	w := httptest.NewRecorder()
	newMux().ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader(body)))
	if w.Code != 401 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestVerifyEmptyToken(t *testing.T) {
	body, _ := json.Marshal(verifyRequest{Token: ""})
	w := httptest.NewRecorder()
	newMux().ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestVerifyTokenNoSubject(t *testing.T) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	s, _ := tok.SignedString([]byte(testSecret))
	body, _ := json.Marshal(verifyRequest{Token: s})
	w := httptest.NewRecorder()
	newMux().ServeHTTP(w, httptest.NewRequest("POST", "/verify", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
}
