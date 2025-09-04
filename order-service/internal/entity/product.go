package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Product represents a product in the marketplace.
type Product struct {
	ID               uuid.UUID
	Name             string
	UnitPrice        decimal.Decimal
	Quantity         int64
	Version          int64 // for optimistic locking
	ReservedQuantity int64 // quantity reserved for orders
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// GetAvailableStock returns the available stock (quantity - reserved).
func (p *Product) GetAvailableStock() int64 {
	return p.Quantity - p.ReservedQuantity
}
