// Package httperror provides custom error responses for the product service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-template/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-template/services/product-service/internal/constant"
)

// NewInvalidRequestBodyError creates a new invalid request body error.
func NewInvalidRequestBodyError() *httperror.ResponseError {
	msg := constant.InvalidRequestBodyErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}

// NewInvalidProductIDError creates a new invalid product ID error.
func NewInvalidProductIDError() *httperror.ResponseError {
	msg := constant.InvalidProductIDErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}

// NewNameRequiredError creates a new name required error.
func NewNameRequiredError() *httperror.ResponseError {
	msg := constant.NameRequiredErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}

// NewPriceMustBeGreaterThanZeroError creates a new price validation error.
func NewPriceMustBeGreaterThanZeroError() *httperror.ResponseError {
	msg := constant.PriceMustBeGreaterThanZeroErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}
