package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
)

// NewUnauthorizedError creates a new ResponseError for unauthorized access.
func NewUnauthorizedError() *ResponseError {
	msg := constant.UnauthorizedErrorMessage
	err := errors.New(msg)

	return NewResponseError(err, http.StatusUnauthorized, msg)
}
