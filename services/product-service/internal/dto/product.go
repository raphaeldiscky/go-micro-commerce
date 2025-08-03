// Package dto contains data transfer objects for product service.
package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CreateProductRequest represents the request to create a new product.
type CreateProductRequest struct {
	Name     string          `json:"name"     validate:"required,min=1,max=255"`
	Price    decimal.Decimal `json:"price"    validate:"required,decimal_gt"`
	Quantity int             `json:"quantity" validate:"required,min=0"`
}

// UpdateProductRequest represents the request to update a product.
type UpdateProductRequest struct {
	ID       uuid.UUID       `json:"id"       validate:"required"`
	Name     string          `json:"name"     validate:"required,min=1,max=255"`
	Price    decimal.Decimal `json:"price"    validate:"required,decimal_gt"`
	Quantity int             `json:"quantity" validate:"required,min=0"`
}

// ProductResponse represents a product in API responses.
type ProductResponse struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Quantity  int             `json:"quantity"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// ProductListResponse represents a list of products.
type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// GetProductsRequest represents pagination and filtering parameters.
type GetProductsRequest struct {
	Limit  int `json:"limit"  validate:"min=1,max=100"`
	Offset int `json:"offset" validate:"min=0"`
}
