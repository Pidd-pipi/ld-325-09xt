package errors

import (
	"errors"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid request input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("conflicting request state")
)

type BusinessError struct {
	Code    int
	Message string
	Err     error
}

func (e *BusinessError) Error() string { return e.Message }
func (e *BusinessError) Unwrap() error { return e.Err }

// NewBusinessError wraps a sentinel error with an API-facing code and message.
func NewBusinessError(code int, message string, cause error) *BusinessError {
	return &BusinessError{Code: code, Message: message, Err: cause}
}

// Conflict builds a 409-style business error around ErrConflict.
func Conflict(message string) *BusinessError {
	return NewBusinessError(constants.ErrorConflict, message, ErrConflict)
}
