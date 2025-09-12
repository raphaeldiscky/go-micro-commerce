// Package constant defines error messages used throughout the payment service.
package constant

const (
	// FailedToPingDatabase indicates an error when the database connection cannot be established.
	FailedToPingDatabase = "failed to ping database: %w"
	// PaymentNotFoundErrorMessage is the message returned when a paymentis not found.
	PaymentNotFoundErrorMessage = "payment not found"
	// InboxEventNotFoundErrorMessage is the message returned when an inbox event is not found.
	InboxEventNotFoundErrorMessage = "inbox event not found"
	// OutboxEventNotFoundErrorMessage is the message returned when an outbox event is not found.
	OutboxEventNotFoundErrorMessage = "outbox event not found"
	// InvalidRequestBodyErrorMessage is the message returned when request body is invalid.
	InvalidRequestBodyErrorMessage = "invalid request body"
	// InvalidPaymentIDErrorMessage is the message returned when paymentID is invalid.
	InvalidPaymentIDErrorMessage = "invalid paymentID"
	// NameRequiredErrorMessage is the message returned when payment name is missing.
	NameRequiredErrorMessage = "name is required"
	// PriceMustBeGreaterThanZeroErrorMessage is the message returned when price is invalid.
	PriceMustBeGreaterThanZeroErrorMessage = "price must be greater than 0"
)

const (
	// ProductNotFoundErrorMessage is the message returned when a product is not found.
	ProductNotFoundErrorMessage = "product not found"
	// InsufficientProductStockErrorMessage is the message returned when there is insufficient product stock.
	InsufficientProductStockErrorMessage = "insufficient product stock"
)
