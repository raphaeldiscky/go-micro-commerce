package dto

import "github.com/google/uuid"

// ReserveProductsRequest represents the request to reserve products for an order.
type ReserveProductsRequest struct {
	IdempotencyKey string                   `json:"idempotency_key" validate:"required"`
	Items          []ProductReservationItem `json:"items"           validate:"required,dive"`
}

// ProductReservationItem represents a single product reservation.
type ProductReservationItem struct {
	ProductID       uuid.UUID `json:"product_id"       validate:"required"`
	Quantity        int64     `json:"quantity"         validate:"required,min=1"`
	ExpectedVersion int64     `json:"expected_version" validate:"required,min=1"`
}

// ReleaseProductsRequest represents the request to release reserved products.
type ReleaseProductsRequest struct {
	Items []ProductReservationItem `json:"items" validate:"required,dive"`
}

// ConfirmProductsDeductionRequest represents the request to confirm product deduction.
type ConfirmProductsDeductionRequest struct {
	Items []ProductReservationItem `json:"items" validate:"required,dive"`
}

// RestoreProductsRequest represents the request to restore products.
type RestoreProductsRequest struct {
	Items  []ProductRestorationItem `json:"items"  validate:"required,dive"`
	Reason string                   `json:"reason" validate:"required"`
}

// ProductRestorationItem represents a single product restoration.
type ProductRestorationItem struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int64     `json:"quantity"   validate:"required,min=1"`
}
