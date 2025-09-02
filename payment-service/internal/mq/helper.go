// Package mq defines domain events for the product service.
package mq

import (
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

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
