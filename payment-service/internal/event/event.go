// Package event defines domain events for the product service.
package event

import (
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/constant"
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

// mapStatusToEventType maps payment status to Kafka event type.
func mapStatusToEventType(status constant.PaymentStatus) string {
	switch status {
	case constant.PaymentStatusPending:
		return constant.KafkaEventTypePaymentCreated
	case constant.PaymentStatusConfirmed:
		return constant.KafkaEventTypePaymentConfirmed
	case constant.PaymentStatusPaid:
		return constant.KafkaEventTypePaymentPaid
	case constant.PaymentStatusShipped:
		return constant.KafkaEventTypePaymentShipped
	case constant.PaymentStatusDelivered:
		return constant.KafkaEventTypePaymentDelivered
	case constant.PaymentStatusCanceled:
		return constant.KafkaEventTypePaymentCanceled
	default:
		return "unknown"
	}
}
