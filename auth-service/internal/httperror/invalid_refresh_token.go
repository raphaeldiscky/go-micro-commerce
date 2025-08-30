package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/constant"
)

// NewInvalidRefreshTokenError is return when refresh token invalid.
func NewInvalidRefreshTokenError() *httperror.ResponseError {
	msg := constant.InvalidCredentialErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}
