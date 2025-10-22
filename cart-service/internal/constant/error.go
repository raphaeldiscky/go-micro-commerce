// Package constant defines error messages used throughout the cart service.
package constant

const (
	// FailedToPingDatabase indicates an error when the database connection cannot be established.
	FailedToPingDatabase = "failed to ping database: %w"
	// CartNotFoundErrorMessage is the message returned when a cart is not found.
	CartNotFoundErrorMessage = "cart not found"
	// InvalidRequestBodyErrorMessage is the message returned when request body is invalid.
	InvalidRequestBodyErrorMessage = "invalid request body"
	// InvalidCartIDErrorMessage is the message returned when cart ID is invalid.
	InvalidCartIDErrorMessage = "invalid cart ID"
	// NameRequiredErrorMessage is the message returned when cart name is missing.
	NameRequiredErrorMessage = "name is required"
	// PriceMustBeGreaterThanZeroErrorMessage is the message returned when price is invalid.
	PriceMustBeGreaterThanZeroErrorMessage = "price must be greater than 0"
	// CheckoutSessionNotFoundErrorMessage is the message returned when a checkout session is not found.
	CheckoutSessionNotFoundErrorMessage = "checkout session not found"
)

const (
	// ProductNotFoundErrorMessage is the message returned when a product is not found.
	ProductNotFoundErrorMessage = "product not found"
	// InsufficientProductStockErrorMessage is the message returned when there is insufficient product stock.
	InsufficientProductStockErrorMessage = "insufficient product stock"
)

const (
	// InboxEventNotFoundErrorMessage is the message returned when an inbox event is not found.
	InboxEventNotFoundErrorMessage = "inbox event not found"
	// OutboxEventNotFoundErrorMessage is the message returned when an outbox event is not found.
	OutboxEventNotFoundErrorMessage = "outbox event not found"
	// SagaStateNotFoundErrorMessage is the message returned when a saga state is not found.
	SagaStateNotFoundErrorMessage = "saga state not found"
)
