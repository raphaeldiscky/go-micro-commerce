package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Product represents a product in the marketplace.
type Product struct {
	ID               uuid.UUID       `json:"id"`
	Name             string          `json:"name"`
	UnitPrice        decimal.Decimal `json:"unit_price"`
	Quantity         int64           `json:"quantity"`
	Version          int64           `json:"version"`           // for optimistic locking
	ReservedQuantity int64           `json:"reserved_quantity"` // quantity reserved for carts
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// GetAvailableStock returns the available stock (quantity - reserved).
func (p *Product) GetAvailableStock() int64 {
	return p.Quantity - p.ReservedQuantity
}
