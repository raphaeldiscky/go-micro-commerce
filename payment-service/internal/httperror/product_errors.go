package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-template/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/constant"
)

// NewProductNotFoundError creates a new Product not found error.
func NewProductNotFoundError() *httperror.ResponseError {
	msg := constant.ProductNotFoundErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusNotFound, msg)
}
