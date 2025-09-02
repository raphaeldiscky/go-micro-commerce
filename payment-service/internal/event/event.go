// Package event defines domain events for the product service.
package event

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

type (
	// BaseEvent defines the interface for all events in the product service.
	BaseEvent = kafka.BaseEvent
	// KafkaMetadata provides common event properties.
	KafkaMetadata = kafka.Metadata
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
	case constant.PaymentStatusProcessing:
		return constant.KafkaEventTypePaymentProcessing
	case constant.PaymentStatusCompleted:
		return constant.KafkaEventTypePaymentCompleted
	case constant.PaymentStatusFailed:
		return constant.KafkaEventTypePaymentFailed
	case constant.PaymentStatusRefunded:
		return constant.KafkaEventTypePaymentRefunded
	default:
		return "unknown"
	}
}
