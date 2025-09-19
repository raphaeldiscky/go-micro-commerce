package entity

import "github.com/google/uuid"

// ProductReservation represents a product reservation request.
type ProductReservation struct {
	ProductID       uuid.UUID
	Quantity        int64
	ExpectedVersion int64
}

// ProductRestoration represents a product restoration/release request without version requirements.
type ProductRestoration struct {
	ProductID uuid.UUID
	Quantity  int64
}
