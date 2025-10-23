package mapper

import (
	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
)

// MapToGraphQLCartFromDTO maps dto.CartResponse to graph.Cart.
func MapToGraphQLCartFromDTO(cart *dto.CartResponse) *graph.Cart {
	items := make([]*graph.CartItem, len(cart.Items))
	for i := range cart.Items {
		items[i] = MapToGraphQLCartItemFromDTO(&cart.Items[i])
	}

	return &graph.Cart{
		ID:         cart.ID,
		CustomerID: cart.CustomerID,
		Status:     cart.Status,
		Items:      items,
		CreatedAt:  cart.CreatedAt,
		UpdatedAt:  cart.UpdatedAt,
	}
}

// MapToGraphQLCartItemFromDTO maps dto.CartItemResponse to graph.CartItem.
func MapToGraphQLCartItemFromDTO(item *dto.CartItemResponse) *graph.CartItem {
	return &graph.CartItem{
		ID:                  item.ID,
		CartID:              item.ID, // CartID will be set by the parent cart
		ProductID:           item.ProductID,
		Quantity:            int(item.Quantity),
		SelectedForCheckout: item.SelectedForCheckout,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
}

// MapToAddCartItemRequest maps graph.AddCartItemInput to dto.AddCartItemRequest.
func MapToAddCartItemRequest(
	input graph.AddCartItemInput,
	customerID uuid.UUID,
) *dto.AddCartItemRequest {
	return &dto.AddCartItemRequest{
		CustomerID: customerID,
		ProductID:  input.ProductID,
		Quantity:   int64(input.Quantity),
	}
}
