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
