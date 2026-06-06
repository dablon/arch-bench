// Package main is a FLAT-architecture JWT verifier.
//
// Architectural style: FLAT (a.k.a. "transaction script").
// All code lives in this single file. There is one function per concern
// (verify, health, main) and no abstractions. The wire format, the JWT
// logic, the HTTP routing, the JSON encoding, the env config — all
// visible in 200 lines of straight-line code.
//
// Pro: minimum moving parts, easiest to read for someone landing on
// the code for the first time.
// Con: when this grows past ~500 lines (multiple endpoints, multiple
// token formats, multiple transports, tests, observability), the lack
// of seams makes changes painful.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type verifyRequest struct {
	Token string `json:"token"`
}

type verifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject,omitempty"`
	Code    string `json:"code"`
}

func verifyToken(token, secret string) (sub string, ok bool, code string) {
	if token == "" {
		return "", false, "ERR_BAD_REQUEST"
	}
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("bad alg")
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return "", false, "ERR_INVALID_TOKEN"
	}
	claims, _ := parsed.Claims.(jwt.MapClaims)
	sub, _ = claims["sub"].(string)
	return sub, true, "OK"
}

func handleVerify(w http.ResponseWriter, r *http.Request, secret string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var req verifyRequest
	if err := json.Unmarshal(body, &req); err != nil || req.Token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(verifyResponse{Valid: false, Code: "ERR_BAD_REQUEST"})
		return
	}
	sub, ok, code := verifyToken(req.Token, secret)
	status := http.StatusOK
	if !ok {
		status = http.StatusUnauthorized
		if code == "ERR_BAD_REQUEST" {
			status = http.StatusBadRequest
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(verifyResponse{Valid: ok, Subject: sub, Code: code})
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	if secret == "" {
		log.Fatal("JWT_SECRET required")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		handleVerify(w, r, secret)
	})
	mux.HandleFunc("/health", handleHealth)
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("go-flat-http listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}
