package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// NewRequestDuplicateError creates a new ResponseError for duplicate requests.
func NewRequestDuplicateError() *ResponseError {
	msg := constant.RequestDuplicateErrorMessage
	err := errors.New(msg)

	return NewResponseError(err, http.StatusConflict, msg)
}
