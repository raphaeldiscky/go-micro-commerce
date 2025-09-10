// Package httperror provides custom error responses for the product service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
)

// NewInvalidRequestBodyError creates a new invalid request body error.
func NewInvalidRequestBodyError() *httperror.ResponseError {
	msg := constant.InvalidRequestBodyErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}
