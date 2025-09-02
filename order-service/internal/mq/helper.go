// Package mq defines domain events for the product service.
package mq

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event/payload"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// mapStatusToEventType maps order status to Kafka event type.
func mapStatusToEventType(status constant.OrderStatus) string {
	switch status {
	case constant.OrderStatusPending:
		return event.OrderCreatedEventType
	case constant.OrderStatusPaid:
		return event.OrderPaidEventType
	case constant.OrderStatusShipped:
		return event.OrderShippedEventType
	case constant.OrderStatusDelivered:
		return event.OrderDeliveredEventType
	case constant.OrderStatusCanceled:
		return event.OrderCanceledEventType
	case constant.OrderStatusConfirmed:
		return event.OrderConfirmedEventType
	case constant.OrderStatusProcessing:
		return event.OrderProcessingEventType
	case constant.OrderStatusFailed:
		return event.OrderFailedEventType
	default:
		return "unknown"
	}
}

// mapOrderItemsToPayload maps order items to their payload representation.
func mapOrderItemsToPayload(items []entity.OrderItem) []payload.OrderItemPayload {
	payloadItems := make([]payload.OrderItemPayload, len(items))

	for i := range items {
		item := &items[i]
		payloadItems[i] = payload.OrderItemPayload{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return payloadItems
}
