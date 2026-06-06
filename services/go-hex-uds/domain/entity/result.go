// Package entity holds the value objects of the UDS domain.
package entity

type ResultCode string

const (
	CodeOK           ResultCode = "OK"
	CodeBadRequest   ResultCode = "ERR_BAD_REQUEST"
	CodeInvalidToken ResultCode = "ERR_INVALID_TOKEN"
)

type VerificationResult struct {
	Valid   bool
	Subject string
	Code    ResultCode
}
