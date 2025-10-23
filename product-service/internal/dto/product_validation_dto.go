package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ValidateProductsRequest represents the request to validate products before checkout.
type ValidateProductsRequest struct {
	Products []ProductValidationItem
}

// ProductValidationItem represents a product to validate.
type ProductValidationItem struct {
	ID       uuid.UUID
	Price    decimal.Decimal
	Quantity int64
}
