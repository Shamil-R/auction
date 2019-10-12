package errors

import (
	"errors"
	"net/http"
)

type CodeError interface {
	Code() int
}

type codeError struct {
	Message string `json:"message"`
	code    int
}

func (e *codeError) Error() string {
	return e.Message
}

func (e *codeError) Code() int {
	return e.code
}

func New(message string) error {
	return errors.New(message)
}

func NewError(message string, code int) *codeError {
	return &codeError{Message: message, code: code}
}

func BadRequest(message string) *codeError {
	return NewError(message, http.StatusBadRequest)
}

func Forbidden(message string) *codeError {
	return NewError(message, http.StatusForbidden)
}

func NotFound(message string) *codeError {
	return NewError(message, http.StatusNotFound)
}
