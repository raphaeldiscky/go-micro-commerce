// Package httperror provides custom error responses for the Payment service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// NewPaymentNotFoundError creates a new Payment not found error.
func NewPaymentNotFoundError() *httperror.ResponseError {
	msg := constant.PaymentNotFoundErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusNotFound, msg)
}

// NewInsufficientProductStockError creates a new Insufficient Product Stock error.
func NewInsufficientProductStockError() *httperror.ResponseError {
	msg := constant.InsufficientProductStockErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusConflict, msg)
}
