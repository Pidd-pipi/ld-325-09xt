package errors

import (
	"errors"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrInvalidInput = errors.New("invalid request input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("resource conflict")
)

type BusinessError struct {
	Code    int
	Message string
	Err     error
}

func (e *BusinessError) Error() string { return e.Message }
func (e *BusinessError) Unwrap() error { return e.Err }

// NewConflictError 标记一个携带可读说明的业务冲突（HTTP 409）。
func NewConflictError(message string) *BusinessError {
	return &BusinessError{Message: message, Err: ErrConflict}
}

// NewNotFoundError 标记一个携带可读说明的资源不存在（HTTP 404）。
func NewNotFoundError(message string) *BusinessError {
	return &BusinessError{Message: message, Err: ErrNotFound}
}

// NewBusinessValidation 标记一条面向用户的业务规则校验失败（HTTP 400）。
func NewBusinessValidation(message string) *BusinessError {
	return &BusinessError{Code: constants.ErrorValidation, Message: message, Err: ErrInvalidInput}
}
