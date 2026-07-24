package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"error"`
	Detail  string `json:"detail,omitempty"`
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Detail)
	}
	return e.Message
}

func (e *AppError) StatusCode() int {
	return e.Code
}

func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func Newf(code int, format string, args ...interface{}) *AppError {
	return &AppError{Code: code, Message: fmt.Sprintf(format, args...)}
}

func Wrap(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Detail: err.Error()}
}

var (
	ErrBadRequest      = func(msg string) *AppError { return New(http.StatusBadRequest, msg) }
	ErrNotFound        = func(msg string) *AppError { return New(http.StatusNotFound, msg) }
	ErrInternal        = func(msg string) *AppError { return New(http.StatusInternalServerError, msg) }
	ErrInternalWrap    = func(msg string, err error) *AppError { return Wrap(http.StatusInternalServerError, msg, err) }
	ErrTooManyRequests = func() *AppError { return New(http.StatusTooManyRequests, "rate limit exceeded") }
	ErrServiceTimeout  = func() *AppError { return New(http.StatusServiceUnavailable, "service temporarily unavailable") }
)
