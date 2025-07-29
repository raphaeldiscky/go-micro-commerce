// Package common defines the SellerResult structure used in application responses.
package common

import (
	"time"

	"github.com/google/uuid"
)

// SellerResult represents the result of a seller query.
type SellerResult struct {
	ID        uuid.UUID
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
