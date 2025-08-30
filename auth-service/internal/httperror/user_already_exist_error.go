package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
)

// NewUserAlreadyExistError is returned when a user already exists.
func NewUserAlreadyExistError() *httperror.ResponseError {
	msg := constant.UserAlreadyExistErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}
