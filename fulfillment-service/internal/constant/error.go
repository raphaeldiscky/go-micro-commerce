// Package constant defines error messages used throughout the fulfillment service.
package constant

const (
	// FailedToPingDatabase indicates an error when the database connection cannot be established.
	FailedToPingDatabase = "failed to ping database: %w"
	// FulfillmentNotFoundErrorMessage is the message returned when a fulfillment is not found.
	FulfillmentNotFoundErrorMessage = "fulfillment not found"
	// InboxEventNotFoundErrorMessage is the message returned when an inbox event is not found.
	InboxEventNotFoundErrorMessage = "inbox event not found"
	// OutboxEventNotFoundErrorMessage is the message returned when an outbox event is not found.
	OutboxEventNotFoundErrorMessage = "outbox event not found"
	// InvalidRequestBodyErrorMessage is the message returned when request body is invalid.
	InvalidRequestBodyErrorMessage = "invalid request body"
	// InvalidFulfillmentIDErrorMessage is the message returned when fulfillmentID is invalid.
	InvalidFulfillmentIDErrorMessage = "invalid fulfillmentID"
	// TrackingNumberRequiredErrorMessage is the message returned when tracking number is missing.
	TrackingNumberRequiredErrorMessage = "tracking number is required"
	// WeightMustBeGreaterThanZeroErrorMessage is the message returned when weight is invalid.
	WeightMustBeGreaterThanZeroErrorMessage = "weight must be greater than 0"
	// FulfillmentAlreadyShippedErrorMessage is the message returned when fulfillment is already shipped.
	FulfillmentAlreadyShippedErrorMessage = "fulfillment already shipped"
	// FulfillmentCannotBeCanceledErrorMessage is the message returned when fulfillment cannot be canceled.
	FulfillmentCannotBeCanceledErrorMessage = "fulfillment cannot be canceled"
)

const (
	// ProductNotFoundErrorMessage is the message returned when a product is not found.
	ProductNotFoundErrorMessage = "product not found"
	// InsufficientProductStockErrorMessage is the message returned when there is insufficient product stock.
	InsufficientProductStockErrorMessage = "insufficient product stock"
)
