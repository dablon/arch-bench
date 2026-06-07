// Package http is the INCOMING adapter that drives the application
// over HTTP.
//
// An "incoming adapter" in hexagonal terms: it's the bridge from the
// outside world to the application's use case. The outside world here
// is HTTP+JSON; in another adapter it could be gRPC, a CLI, a message
// queue, etc.
//
// This adapter's job is mechanical:
//   1. Decode the request.
//   2. Call the use case.
//   3. Encode the result into an HTTP response.
//
// It is intentionally thin. The use case (in domain/) decides the
// outcome; this file just translates.
package transport

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dablon/arch-bench/services/go-hex-http/app"
	"github.com/dablon/arch-bench/services/go-hex-http/domain/entity"
)

type verifyRequest struct {
	Token string `json:"token"`
}

type verifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject,omitempty"`
	Code    string `json:"code"`
}

type Handler struct {
	app *app.Application
}

func NewHandler(application *app.Application) *Handler {
	return &Handler{app: application}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var req verifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, verifyResponse{Valid: false, Code: string(entity.CodeBadRequest)})
		return
	}
	res := h.app.VerifyToken.Execute(req.Token)
	status := http.StatusOK
	if !res.Valid {
		switch res.Code {
		case entity.CodeBadRequest:
			status = http.StatusBadRequest
		default:
			status = http.StatusUnauthorized
		}
	}
	writeJSON(w, status, verifyResponse{Valid: res.Valid, Subject: res.Subject, Code: string(res.Code)})
}

func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
