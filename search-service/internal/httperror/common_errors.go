// Package httperror provides custom error responses for the Search service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"
)

// NewBadRequestError returns a 400 error.
func NewBadRequestError(message string) *httperror.ResponseError {
	err := errors.New(message)

	return httperror.NewResponseError(err, http.StatusBadRequest, message)
}
