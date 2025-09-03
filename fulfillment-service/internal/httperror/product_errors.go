// Package httperror provides custom error responses for the Fulfillment service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
)

// NewInsufficientProductStockError creates a new Insufficient Product Stock error.
func NewInsufficientProductStockError() *httperror.ResponseError {
	msg := constant.InsufficientProductStockErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusConflict, msg)
}
