package entity

import "github.com/google/uuid"

// ProductReservation represents a product reservation request.
type ProductReservation struct {
	ProductID       uuid.UUID
	Quantity        int64
	ExpectedVersion int64
}
