// Package mapper provides functions for mapping domain entities to DTOs and vice versa.
package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/entity"
)

// MapToCartResponse converts domain entity to DTO response.
// Cart is lightweight - only includes items and selection state, no pricing.
func MapToCartResponse(cart *entity.Cart) *dto.CartResponse {
	return &dto.CartResponse{
		ID:         cart.ID,
		CustomerID: cart.CustomerID,
		Items:      MapToCartItemResponses(cart.Items),
		CreatedAt:  cart.CreatedAt,
		UpdatedAt:  cart.UpdatedAt,
	}
}

// MapToCartItemResponses converts domain entities to DTO responses.
// CartItems are lightweight - only include product reference and quantity, no pricing.
func MapToCartItemResponses(items []entity.CartItem) []dto.CartItemResponse {
	if len(items) == 0 {
		return []dto.CartItemResponse{}
	}

	responses := make([]dto.CartItemResponse, 0, len(items))

	for i := range items {
		item := &items[i]
		responses = append(responses, dto.CartItemResponse{
			ID:                  item.ID,
			ProductID:           item.ProductID,
			Quantity:            item.Quantity,
			SelectedForCheckout: item.SelectedForCheckout,
			AddedAt:             item.AddedAt,
		})
	}

	return responses
}
