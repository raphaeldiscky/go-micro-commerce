package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// NewServerError creates a new ResponseError for internal server errors.
func NewServerError() *ResponseError {
	msg := constant.InternalServerErrorMessage
	err := errors.New(msg)

	return NewResponseError(err, http.StatusInternalServerError, msg)
}
