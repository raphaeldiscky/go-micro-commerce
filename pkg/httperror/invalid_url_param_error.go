package httperror

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/raphaeldiscky/go-micro-template/pkg/constant"
)

// NewInvalidURLParamError creates a new ResponseError for invalid URL parameters.
func NewInvalidURLParamError(param string) *ResponseError {
	msg := fmt.Sprintf(constant.InvalidURLParamErrorMessage, param)
	err := errors.New(msg)

	return NewResponseError(err, http.StatusBadRequest, msg)
}
