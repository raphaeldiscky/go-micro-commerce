// Package dto contains data transfer objects for cart service.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateCartItemRequest represents an item in create cart request.
type CreateCartItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int64     `json:"quantity"   validate:"required,min=1"`
}

// CreateCartRequest represents the request to create a new cart.
// Cart is lightweight - only stores items and selection state.
type CreateCartRequest struct {
	CustomerID uuid.UUID               `json:"customer_id"` // from context or header
	Items      []CreateCartItemRequest `json:"items"       validate:"required,min=1,dive"`
}

// AddCartItemRequest represents the request to add an item to a cart.
type AddCartItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int64     `json:"quantity"   validate:"required,min=1"`
}

// UpdateCartItemQuantityRequest represents the request to update an item quantity.
type UpdateCartItemQuantityRequest struct {
	Quantity int64 `json:"quantity" validate:"required,min=1"`
}

// SelectItemForCheckoutRequest represents the request to select/deselect an item for checkout.
type SelectItemForCheckoutRequest struct {
	Selected bool `json:"selected"`
}

// CartItemResponse represents a cart item in API responses.
// Lightweight - no pricing information.
type CartItemResponse struct {
	ID                  uuid.UUID `json:"id"`
	ProductID           uuid.UUID `json:"product_id"`
	Quantity            int64     `json:"quantity"`
	SelectedForCheckout bool      `json:"selected_for_checkout"`
	AddedAt             time.Time `json:"added_at"`
}

// CartResponse represents a cart in API responses.
// Lightweight - pricing calculations happen in CheckoutSession.
type CartResponse struct {
	ID         uuid.UUID          `json:"id"`
	CustomerID uuid.UUID          `json:"customer_id"`
	Items      []CartItemResponse `json:"items"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}
