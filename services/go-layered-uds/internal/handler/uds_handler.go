// Package handler is the UDS-transport layer for the layered
// architecture. It speaks the line protocol and translates service
// results to lines. It knows nothing about JWT, claims, or crypto.
package handler

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/dablon/arch-bench/services/go-layered-uds/internal/service"
)

const maxLine = 4096

type Handler struct {
	svc service.VerifierService
}

func NewHandler(s service.VerifierService) *Handler {
	return &Handler{svc: s}
}

// Handle is called per accepted connection. The connection is closed
// when the read loop returns (EOF or error).
func (h *Handler) Handle(c net.Conn) {
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
			res := h.svc.VerifyToken(tok)
			if res.Valid {
				resp = fmt.Sprintf("OK %s\n", res.Subject)
			} else {
				switch res.Code {
				case service.CodeBadRequest:
					resp = "ERR BAD_REQUEST\n"
				default:
					resp = "ERR INVALID_TOKEN\n"
				}
			}
		}
		_, _ = c.Write([]byte(resp))
	}
}
