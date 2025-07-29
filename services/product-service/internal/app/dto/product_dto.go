package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateProductRequest represents the request to create a new product
type CreateProductRequest struct {
	Name     string    `json:"name" validate:"required,min=1,max=255"`
	Price    float64   `json:"price" validate:"required,gt=0"`
	SellerId uuid.UUID `json:"seller_id" validate:"required"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	Id       uuid.UUID `json:"id" validate:"required"`
	Name     string    `json:"name" validate:"required,min=1,max=255"`
	Price    float64   `json:"price" validate:"required,gt=0"`
	SellerId uuid.UUID `json:"seller_id" validate:"required"`
}

// ProductResponse represents a product in API responses
type ProductResponse struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	SellerId  uuid.UUID `json:"seller_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductListResponse represents a list of products
type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// GetProductsRequest represents pagination and filtering parameters
type GetProductsRequest struct {
	SellerId *uuid.UUID `json:"seller_id,omitempty"`
	Limit    int        `json:"limit" validate:"min=1,max=100"`
	Offset   int        `json:"offset" validate:"min=0"`
}
