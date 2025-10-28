// Package dto contains data transfer objects for order service.
package dto

import "github.com/shopspring/decimal"

// CalculateShippingRequest represents a request to calculate shipping cost.
type CalculateShippingRequest struct {
	Courier     Courier
	Destination ToAddress
	Origin      FromAddress
	Package     Package
	Currency    string
}

// CalculateShippingResponse represents the response from shipping cost calculation.
type CalculateShippingResponse struct {
	Cost     decimal.Decimal
	Currency string
}
