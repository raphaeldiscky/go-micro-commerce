// Package mapper provides functions for mapping domain entities to DTOs and vice versa.
package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// MapToOrderResponse converts domain entity to DTO response.
func MapToOrderResponse(order *entity.Order) *dto.OrderResponse {
	return &dto.OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Status:     order.Status,
		TotalPrice: order.TotalPrice,
		Items:      MapToOrderItemResponses(order.Items),
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}

// MapToOrderItemResponses converts domain entities to DTO responses.
func MapToOrderItemResponses(items []entity.OrderItem) []dto.OrderItemResponse {
	var responses []dto.OrderItemResponse

	for i := range items {
		item := &items[i]
		responses = append(responses, dto.OrderItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	return responses
}
