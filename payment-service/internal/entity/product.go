package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Product represents a product in the marketplace.
type Product struct {
	ID        uuid.UUID
	Name      string
	Price     decimal.Decimal
	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}
