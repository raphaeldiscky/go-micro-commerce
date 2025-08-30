// Package httperror provides custom error responses for the Order service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// NewOrderNotFoundError creates a new Order not found error.
func NewOrderNotFoundError() *httperror.ResponseError {
	msg := constant.OrderNotFoundErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusNotFound, msg)
}

// NewInsufficientProductStockError creates a new Insufficient Product Stock error.
func NewInsufficientProductStockError() *httperror.ResponseError {
	msg := constant.InsufficientProductStockErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusConflict, msg)
}
