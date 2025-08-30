package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Product represents a product in the marketplace.
type Product struct {
	ID                uuid.UUID
	Name              string
	Price             decimal.Decimal
	Quantity          int
	Version           int64 // for optimistic locking
	AllocatedQuantity int   // quantity reserved for orders
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// GetAvailableStock returns the available stock (quantity - allocated).
func (p *Product) GetAvailableStock() int {
	return p.Quantity - p.AllocatedQuantity
}
