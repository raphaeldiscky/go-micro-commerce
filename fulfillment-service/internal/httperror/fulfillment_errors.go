// Package httperror provides custom error responses for the Fulfillment service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
)

// NewFulfillmentNotFoundError creates a new Fulfillment not found error.
func NewFulfillmentNotFoundError() *httperror.ResponseError {
	msg := constant.FulfillmentNotFoundErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusNotFound, msg)
}

// NewFulfillmentAlreadyShippedError creates a new Fulfillment already shipped error.
func NewFulfillmentAlreadyShippedError() *httperror.ResponseError {
	msg := constant.FulfillmentAlreadyShippedErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusConflict, msg)
}

// NewFulfillmentCannotBeCanceledError creates a new Fulfillment cannot be canceled error.
func NewFulfillmentCannotBeCanceledError() *httperror.ResponseError {
	msg := constant.FulfillmentCannotBeCanceledErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusConflict, msg)
}
