package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"
)

// NewNotFoundError returns a 404 error.
func NewNotFoundError(message string) *httperror.ResponseError {
	err := errors.New(message)

	return httperror.NewResponseError(err, http.StatusNotFound, message)
}

// NewForbiddenError returns a 403 error.
func NewForbiddenError(message string) *httperror.ResponseError {
	err := errors.New(message)

	return httperror.NewResponseError(err, http.StatusForbidden, message)
}

// NewBadRequestError returns a 400 error.
func NewBadRequestError(message string) *httperror.ResponseError {
	err := errors.New(message)

	return httperror.NewResponseError(err, http.StatusBadRequest, message)
}

// NewInternalServerError is return when email or password wrong.
func NewInternalServerError(message string) *httperror.ResponseError {
	err := errors.New(message)

	return httperror.NewResponseError(err, http.StatusInternalServerError, message)
}

// NewUnauthorizedError returns a 401 error.
func NewUnauthorizedError(message string) *httperror.ResponseError {
	err := errors.New(message)

	return httperror.NewResponseError(err, http.StatusUnauthorized, message)
}
