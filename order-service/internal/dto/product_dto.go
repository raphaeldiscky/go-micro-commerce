package dto

import "github.com/google/uuid"

// ProductReservationItem represents a product reservation request.
type ProductReservationItem struct {
	ProductID       uuid.UUID
	Quantity        int64
	ExpectedVersion int64
}

// ProductRestorationItem represents a product restoration request.
type ProductRestorationItem struct {
	ProductID uuid.UUID
	Quantity  int64
}
