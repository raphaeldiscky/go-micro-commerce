// Package httperror is a custom error package for handling HTTP errors.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-template/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/constant"
)

// NewInvalidCredentialError is return when email or password wrong.
func NewInvalidCredentialError() *httperror.ResponseError {
	msg := constant.InvalidCredentialErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}
