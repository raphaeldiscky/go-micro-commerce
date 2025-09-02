package event

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductCreatedPayload holds the data for the product created event.
type ProductCreatedPayload struct {
	ProductID uuid.UUID       `json:"product_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Quantity  int64           `json:"quantity"`
}

// ProductDeletedPayload represents when a product is deleted.
type ProductDeletedPayload struct {
	ProductID uuid.UUID `json:"product_id"`
}

// ProductUpdatedPayload holds the data for the product updated event.
type ProductUpdatedPayload struct {
	ProductID uuid.UUID       `json:"product_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Quantity  int64           `json:"quantity"`
}
