// Package handler is the HTTP-transport layer of the layered architecture.
//
// The handler:
//   - parses the request (path, method, body, JSON)
//   - calls the service layer
//   - maps the service's domain Result to an HTTP response (status code + JSON)
//
// The handler knows nothing about JWT or claims.
package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/dablon/arch-bench/services/go-layered-http/internal/service"
)

type verifyRequest struct {
	Token string `json:"token"`
}

type verifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject,omitempty"`
	Code    string `json:"code"`
}

type VerifyHandler struct {
	svc service.VerifierService
}

func NewVerifyHandler(s service.VerifierService) *VerifyHandler {
	return &VerifyHandler{svc: s}
}

func (h *VerifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var req verifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, verifyResponse{Valid: false, Code: string(service.CodeBadRequest)})
		return
	}
	res := h.svc.VerifyToken(req.Token)
	status := http.StatusOK
	if !res.Valid {
		switch res.Code {
		case service.CodeBadRequest:
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
