// Package mapper provides functions for mapping domain entities to DTOs and vice versa.
package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// MapToOrderResponse converts domain entity to DTO response.
func MapToOrderResponse(order *entity.Order) *dto.OrderResponse {
	return &dto.OrderResponse{
		ID:                order.ID,
		IdempotencyKey:    order.IdempotencyKey,
		CheckoutSessionID: order.CheckoutSessionID,
		CustomerID:        order.CustomerID,
		Status:            order.Status,
		Currency:          order.Currency,
		PaymentGateway:    order.PaymentGateway,
		Courier: dto.Courier{
			CourierID: order.Courier.CourierID,
		},
		Package: dto.Package{
			WeightKG: order.Package.WeightKG,
			Length:   order.Package.Length,
			Height:   order.Package.Height,
			Width:    order.Package.Width,
			Unit:     order.Package.Unit,
		},
		Origin: dto.FromAddress{
			City:       order.Origin.City,
			State:      order.Origin.State,
			PostalCode: order.Origin.PostalCode,
			Country:    order.Origin.Country,
		},
		Destination: dto.ToAddress{
			City:       order.Destination.City,
			State:      order.Destination.State,
			PostalCode: order.Destination.PostalCode,
			Country:    order.Destination.Country,
		},
		ShippingCost:  order.ShippingCost,
		Subtotal:      order.Subtotal,
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
			OrderID:       item.OrderID,
			ProductID:     item.ProductID,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			TaxRate:       item.TaxRate,
			TotalPrice:    item.TotalPrice,
			TotalTax:      item.TotalTax,
			TotalDiscount: item.TotalDiscount,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
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
	case constant.OrderStatusPaymentPending:
		return kafka.OrderPaymentPendingEventType
	case constant.OrderStatusPaymentExpired:
		return kafka.OrderPaymentExpiredEventType
	case constant.OrderStatusProcessing:
		return kafka.OrderProcessingEventType
	case constant.OrderStatusFailed:
		return kafka.OrderFailedEventType
	case constant.OrderStatusCompleted:
		return kafka.OrderCompletedEventType
	default:
		return "unknown"
	}
}

// MapOrderItemsToPayload maps order items to their payload representation.
func MapOrderItemsToPayload(items []entity.OrderItem) []kafkaevent.OrderItemPayload {
	payloadItems := make([]kafkaevent.OrderItemPayload, len(items))

	for i := range items {
		item := &items[i]
		payloadItems[i] = kafkaevent.OrderItemPayload{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return payloadItems
}
