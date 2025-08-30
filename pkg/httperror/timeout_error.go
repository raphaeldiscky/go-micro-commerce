package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// NewTimeoutError creates a new ResponseError for request timeout.
func NewTimeoutError() *ResponseError {
	msg := constant.RequestTimeoutErrorMessage
	err := errors.New(msg)

	return NewResponseError(err, http.StatusRequestTimeout, msg)
}
