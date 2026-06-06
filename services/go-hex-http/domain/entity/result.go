// Package entity contains the value objects of the domain.
//
// These are framework-free: no JSON tags, no protocol awareness, no
// crypto types. They are the canonical "what the system means" types.
package entity

// ResultCode is the canonical outcome of a verification, expressed as
// a string so it can travel through any transport unchanged.
type ResultCode string

const (
	CodeOK           ResultCode = "OK"
	CodeBadRequest   ResultCode = "ERR_BAD_REQUEST"
	CodeInvalidToken ResultCode = "ERR_INVALID_TOKEN"
)

// VerificationResult is what the system returns to the caller. Valid
// is true only on CodeOK. Subject is the JWT's "sub" claim, empty
// otherwise.
type VerificationResult struct {
	Valid   bool
	Subject string
	Code    ResultCode
}
