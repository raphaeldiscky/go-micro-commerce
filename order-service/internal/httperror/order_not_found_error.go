// Package httperror provides custom error responses for the Order service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-template/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/constant"
)

// NewOrderNotFoundError creates a new Order not found error.
func NewOrderNotFoundError() *httperror.ResponseError {
	msg := constant.OrderNotFoundErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusNotFound, msg)
}
