// Package mq defines domain events for the product service.
package mq

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event/payload"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// mapStatusToEventType maps order status to Kafka event type.
func mapStatusToEventType(status constant.OrderStatus) string {
	switch status {
	case constant.OrderStatusPending:
		return constant.KafkaEventTypeOrderCreated
	case constant.OrderStatusPaid:
		return constant.KafkaEventTypeOrderPaid
	case constant.OrderStatusShipped:
		return constant.KafkaEventTypeOrderShipped
	case constant.OrderStatusDelivered:
		return constant.KafkaEventTypeOrderDelivered
	case constant.OrderStatusCanceled:
		return constant.KafkaEventTypeOrderCanceled
	case constant.OrderStatusConfirmed:
		return constant.KafkaEventTypeOrderConfirmed
	case constant.OrderStatusProcessing:
		return constant.KafkaEventTypeOrderProcessing
	case constant.OrderStatusFailed:
		return constant.KafkaEventTypeOrderFailed
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
