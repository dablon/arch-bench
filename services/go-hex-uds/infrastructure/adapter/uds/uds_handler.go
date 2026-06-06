// Package uds is the INCOMING adapter that drives the application
// over a Unix Domain Socket using the line protocol.
package uds

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/dablon/arch-bench/services/go-hex-uds/domain/entity"
	"github.com/dablon/arch-bench/services/go-hex-uds/domain/usecase"
)

const maxLine = 4096

type Handler struct{ uc *usecase.VerifyToken }

func NewHandler(uc *usecase.VerifyToken) *Handler { return &Handler{uc: uc} }

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
			res := h.uc.Execute(tok)
			if res.Valid {
				resp = fmt.Sprintf("OK %s\n", res.Subject)
			} else {
				switch res.Code {
				case entity.CodeBadRequest:
					resp = "ERR BAD_REQUEST\n"
				default:
					resp = "ERR INVALID_TOKEN\n"
				}
			}
		}
		_, _ = c.Write([]byte(resp))
	}
}
