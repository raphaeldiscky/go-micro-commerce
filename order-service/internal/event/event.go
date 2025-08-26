// Package event defines domain events for the product service.
package event

import (
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/entity"
)

type (
	// BaseEvent defines the interface for all events in the product service.
	BaseEvent = mq.BaseEvent
	// KafkaMetadata provides common event properties.
	KafkaMetadata = mq.KafkaMetadata
)

// AsyncProducer defines the interface for producing events.
type AsyncProducer interface {
	ProduceAsync(topic string, event BaseEvent) error
}

// mapStatusToEventType maps order status to Kafka event type.
func mapStatusToEventType(status constant.OrderStatus) string {
	switch status {
	case constant.OrderStatusPending:
		return constant.KafkaEventTypeOrderCreated
	case constant.OrderStatusConfirmed:
		return constant.KafkaEventTypeOrderConfirmed
	case constant.OrderStatusPaid:
		return constant.KafkaEventTypeOrderPaid
	case constant.OrderStatusShipped:
		return constant.KafkaEventTypeOrderShipped
	case constant.OrderStatusDelivered:
		return constant.KafkaEventTypeOrderDelivered
	case constant.OrderStatusCanceled:
		return constant.KafkaEventTypeOrderCanceled
	default:
		return "unknown"
	}
}

// mapOrderItemsToPayload maps order items to their payload representation.
func mapOrderItemsToPayload(items []entity.OrderItem) []OrderItemPayload {
	payloadItems := make([]OrderItemPayload, len(items))
	for i, item := range items {
		payloadItems[i] = OrderItemPayload{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return payloadItems
}
