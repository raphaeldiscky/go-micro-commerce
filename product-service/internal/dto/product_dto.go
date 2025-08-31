// Package dto contains data transfer objects for product service.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/entity"
)

// CreateProductRequest represents the request to create a new product.
type CreateProductRequest struct {
	Name     string          `json:"name"     validate:"required,min=1,max=255"`
	Price    decimal.Decimal `json:"price"    validate:"required,decimal_gte"`
	Quantity int64           `json:"quantity" validate:"required,min=0"`
}

// UpdateProductRequest represents the request to update a product.
type UpdateProductRequest struct {
	ID       uuid.UUID       `json:"id"       validate:"required"`
	Name     string          `json:"name"     validate:"required,min=1,max=255"`
	Price    decimal.Decimal `json:"price"    validate:"required,decimal_gte"`
	Quantity int64           `json:"quantity" validate:"required,min=0"`
}

// ProductResponse represents a product in API responses.
type ProductResponse struct {
	ID                uuid.UUID       `json:"id"`
	Name              string          `json:"name"`
	Price             decimal.Decimal `json:"price"`
	Quantity          int64           `json:"quantity"`
	Version           int64           `json:"version"`
	ReservedQuantity  int64           `json:"reserved_quantity"`
	AvailableQuantity int64           `json:"available_quantity"` // computed field
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// GetProductsRequest represents pagination and filtering parameters.
type GetProductsRequest struct {
	Limit int64 `json:"limit" validate:"min=1,max=100"`
	Page  int64 `json:"page"  validate:"min=1"`
}

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

// MapToProductResponse converts domain entity to DTO response.
func MapToProductResponse(product *entity.Product) *ProductResponse {
	return &ProductResponse{
		ID:                product.ID,
		Name:              product.Name,
		Price:             product.Price,
		Quantity:          product.Quantity,
		Version:           product.Version,
		ReservedQuantity:  product.ReservedQuantity,
		AvailableQuantity: product.GetAvailableStock(),
		CreatedAt:         product.CreatedAt,
		UpdatedAt:         product.UpdatedAt,
	}
}
