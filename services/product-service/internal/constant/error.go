// Package constant defines error messages used throughout the product service.
package constant

const (
	// FailedToPingDatabase indicates an error when the database connection cannot be established.
	FailedToPingDatabase = "failed to ping database: %w"
	// ProductNotFoundErrorMessage is the message returned when a product is not found.
	ProductNotFoundErrorMessage = "product not found"
	// InvalidRequestBodyErrorMessage is the message returned when request body is invalid.
	InvalidRequestBodyErrorMessage = "invalid request body"
	// InvalidProductIDErrorMessage is the message returned when product ID is invalid.
	InvalidProductIDErrorMessage = "invalid product ID"
	// NameRequiredErrorMessage is the message returned when product name is missing.
	NameRequiredErrorMessage = "name is required"
	// PriceMustBeGreaterThanZeroErrorMessage is the message returned when price is invalid.
	PriceMustBeGreaterThanZeroErrorMessage = "price must be greater than 0"
)
