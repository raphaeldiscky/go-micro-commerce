package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateSellerRequest represents the request to create a new seller
type CreateSellerRequest struct {
	Name    string `json:"name" validate:"required,min=2,max=100"`
	Email   string `json:"email" validate:"required,email"`
	Phone   string `json:"phone" validate:"required,min=10,max=20"`
	Address string `json:"address" validate:"required,min=10,max=255"`
}

// UpdateSellerRequest represents the request to update a seller
type UpdateSellerRequest struct {
	Id      uuid.UUID `json:"id"`
	Name    string    `json:"name" validate:"required,min=2,max=100"`
	Email   string    `json:"email" validate:"required,email"`
	Phone   string    `json:"phone" validate:"required,min=10,max=20"`
	Address string    `json:"address" validate:"required,min=10,max=255"`
}

// SellerResponse represents the response for seller operations
type SellerResponse struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetSellersRequest represents the request to get sellers with pagination and filtering
type GetSellersRequest struct {
	Limit      int  `json:"limit"`
	Offset     int  `json:"offset"`
	ActiveOnly bool `json:"active_only"`
}

// SellerListResponse represents the response for listing sellers
type SellerListResponse struct {
	Sellers []SellerResponse `json:"sellers"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// SellerStatusRequest represents the request to change seller status
type SellerStatusRequest struct {
	Id       uuid.UUID `json:"id"`
	IsActive bool      `json:"is_active"`
}
