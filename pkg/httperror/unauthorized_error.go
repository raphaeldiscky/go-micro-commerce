package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// NewUnauthorizedError creates a new ResponseError for unauthorized access.
func NewUnauthorizedError() *ResponseError {
	msg := constant.UnauthorizedErrorMessage
	err := errors.New(msg)

	return NewResponseError(err, http.StatusUnauthorized, msg)
}

// NewMissingXUserIDError creates a new ResponseError for missing user ID.
func NewMissingXUserIDError() *ResponseError {
	msg := constant.MissingXUserIDErrorMessage
	err := errors.New(msg)

	return NewResponseError(err, http.StatusUnauthorized, msg)
}

// NewMissingXEmailError creates a new ResponseError for missing email.
func NewMissingXEmailError() *ResponseError {
	msg := constant.MissingXEmailErrorMessage
	err := errors.New(msg)

	return NewResponseError(err, http.StatusUnauthorized, msg)
}

// NewMissingXRolesError creates a new ResponseError for missing roles.
func NewMissingXRolesError() *ResponseError {
	msg := constant.MissingXRolesErrorMessage
	err := errors.New(msg)

	return NewResponseError(err, http.StatusUnauthorized, msg)
}

// NewMissingXIsActiveError creates a new ResponseError for missing isActive.
func NewMissingXIsActiveError() *ResponseError {
	msg := constant.MissingXIsActiveErrorMessage
	err := errors.New(msg)

	return NewResponseError(err, http.StatusUnauthorized, msg)
}
