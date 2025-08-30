// Package httperror provides custom error responses for the Order service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// NewInvalidRequestBodyError creates a new invalid request body error.
func NewInvalidRequestBodyError() *httperror.ResponseError {
	msg := constant.InvalidRequestBodyErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}

// NewInvalidOrderIDError creates a new invalid Order ID error.
func NewInvalidOrderIDError() *httperror.ResponseError {
	msg := constant.InvalidOrderIDErrorMessage
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
