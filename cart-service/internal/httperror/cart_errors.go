// Package httperror provides custom error responses for the Cart service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
)

// NewCartNotFoundError creates a new Cart not found error.
func NewCartNotFoundError() *httperror.ResponseError {
	msg := constant.CartNotFoundErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusNotFound, msg)
}
