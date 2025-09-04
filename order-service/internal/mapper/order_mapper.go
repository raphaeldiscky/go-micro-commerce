// Package mapper provides functions for mapping domain entities to DTOs and vice versa.
package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// MapToOrderResponse converts domain entity to DTO response.
func MapToOrderResponse(order *entity.Order) *dto.OrderResponse {
	return &dto.OrderResponse{
		ID:            order.ID,
		CustomerID:    order.CustomerID,
		Status:        order.Status,
		Currency:      order.Currency,
		TotalPrice:    order.TotalPrice,
		TotalTax:      order.TotalTax,
		TotalDiscount: order.TotalDiscount,
		Items:         MapToOrderItemResponses(order.Items),
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}
}

// MapToOrderItemResponses converts domain entities to DTO responses.
func MapToOrderItemResponses(items []entity.OrderItem) []dto.OrderItemResponse {
	var responses []dto.OrderItemResponse

	for i := range items {
		item := &items[i]
		responses = append(responses, dto.OrderItemResponse{
			ID:            item.ID,
			ProductID:     item.ProductID,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			TaxRate:       item.TaxRate,
			TotalPrice:    item.TotalPrice,
			TotalTax:      item.TotalTax,
			TotalDiscount: item.TotalDiscount,
		})
	}

	return responses
}

// MapStatusToEventType maps order status to Kafka event type.
func MapStatusToEventType(status constant.OrderStatus) string {
	switch status {
	case constant.OrderStatusPending:
		return kafka.OrderCreatedEventType
	case constant.OrderStatusPaid:
		return kafka.OrderPaidEventType
	case constant.OrderStatusShipped:
		return kafka.OrderShippedEventType
	case constant.OrderStatusDelivered:
		return kafka.OrderDeliveredEventType
	case constant.OrderStatusCanceled:
		return kafka.OrderCanceledEventType
	case constant.OrderStatusConfirmed:
		return kafka.OrderConfirmedEventType
	case constant.OrderStatusProcessing:
		return kafka.OrderProcessingEventType
	case constant.OrderStatusFailed:
		return kafka.OrderFailedEventType
	default:
		return "unknown"
	}
}

// MapOrderItemsToPayload maps order items to their payload representation.
func MapOrderItemsToPayload(items []entity.OrderItem) []event.OrderItemPayload {
	payloadItems := make([]event.OrderItemPayload, len(items))

	for i := range items {
		item := &items[i]
		payloadItems[i] = event.OrderItemPayload{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return payloadItems
}
