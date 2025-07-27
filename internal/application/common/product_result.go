// Package common defines the ProductResult structure used in application responses.
package common

import (
	"time"

	"github.com/google/uuid"
)

// ProductResult represents the result of a product query.
type ProductResult struct {
	Id        uuid.UUID
	Name      string
	Price     float64
	Seller    *SellerResult
	CreatedAt time.Time
	UpdatedAt time.Time
}
