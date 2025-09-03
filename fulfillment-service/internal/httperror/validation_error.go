// Package httperror provides custom error responses for the Order service.
package httperror

import (
	"errors"
	"net/http"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
)

// NewInvalidRequestBodyError creates a new invalid request body error.
func NewInvalidRequestBodyError() *httperror.ResponseError {
	msg := constant.InvalidRequestBodyErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}

// NewInvalidFulfillmentIDError creates a new invalid Fulfillment ID error.
func NewInvalidFulfillmentIDError() *httperror.ResponseError {
	msg := constant.InvalidFulfillmentIDErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}

// NewTrackingNumberRequiredError creates a new tracking number required error.
func NewTrackingNumberRequiredError() *httperror.ResponseError {
	msg := constant.TrackingNumberRequiredErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}

// NewWeightMustBeGreaterThanZeroError creates a new weight validation error.
func NewWeightMustBeGreaterThanZeroError() *httperror.ResponseError {
	msg := constant.WeightMustBeGreaterThanZeroErrorMessage
	err := errors.New(msg)

	return httperror.NewResponseError(err, http.StatusBadRequest, msg)
}
