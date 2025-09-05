package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
)

// MapToOrderSagaResponse converts order response to saga response for async execution.
func MapToOrderSagaResponse(order *dto.OrderResponse) *dto.OrderSagaResponse {
	return &dto.OrderSagaResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Status:     order.Status,
		Currency:   order.Currency,
		Items:      MapToOrderSagaItemResponses(order.Items),
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}

// MapToOrderSagaItemResponses converts order item responses to saga item responses.
func MapToOrderSagaItemResponses(items []dto.OrderItemResponse) []dto.OrderSagaItemResponse {
	var responses []dto.OrderSagaItemResponse

	for i := range items {
		item := &items[i]
		responses = append(responses, dto.OrderSagaItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return responses
}
