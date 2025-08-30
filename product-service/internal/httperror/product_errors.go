// Package httperror provides custom error responses for the product service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
)

// NewProductNotFoundError creates a new product not found error.
func NewProductNotFoundError() *httperror.ResponseError {
	msg := constant.ProductNotFoundErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusNotFound, msg)
}
