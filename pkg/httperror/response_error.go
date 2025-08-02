// Package httperror provides utilities for handling HTTP response errors.
package httperror

import "errors"

// ResponseError represents an error with an associated HTTP status code and message.
type ResponseError struct {
	err  error
	code int
	msg  string
}

// NewResponseError creates a new ResponseError with the provided error, HTTP status code, and message.
func NewResponseError(err error, code int, msg string) *ResponseError {
	return &ResponseError{
		err:  err,
		code: code,
		msg:  msg,
	}
}

// Error implements the error interface for ResponseError.
func (e ResponseError) Error() string {
	if e.msg == "" {
		return e.OriginalMessage()
	}

	return e.msg
}

// GetCode returns the HTTP status code associated with the ResponseError.
func (e ResponseError) GetCode() int {
	return e.code
}

// OriginalError retrieves the original error, traversing through any wrapped ResponseErrors.
func (e ResponseError) OriginalError() error {
	var currErr ResponseError

	currErr = e

	for {
		nextErr := currErr.err
		if nextErr == nil {
			break
		}

		var appErr ResponseError
		if !errors.As(nextErr, &appErr) {
			return nextErr
		}

		currErr = appErr
	}

	return e
}

// OriginalMessage retrieves the original error message from the ResponseError.
func (e ResponseError) OriginalMessage() string {
	return e.OriginalError().Error()
}

// DisplayMessage returns the message associated with the ResponseError.
func (e ResponseError) DisplayMessage() string {
	return e.msg
}
