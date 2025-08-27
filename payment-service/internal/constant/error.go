// Package constant defines error messages used throughout the order service.
package constant

const (
	// FailedToPingDatabase indicates an error when the database connection cannot be established.
	FailedToPingDatabase = "failed to ping database: %w"
	// OrderNotFoundErrorMessage is the message returned when a order is not found.
	OrderNotFoundErrorMessage = "order not found"
	// InvalidRequestBodyErrorMessage is the message returned when request body is invalid.
	InvalidRequestBodyErrorMessage = "invalid request body"
	// InvalidOrderIDErrorMessage is the message returned when order ID is invalid.
	InvalidOrderIDErrorMessage = "invalid order ID"
	// NameRequiredErrorMessage is the message returned when order name is missing.
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
