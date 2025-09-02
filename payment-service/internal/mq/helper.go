// Package mq defines domain events for the product service.
package mq

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// mapStatusToEventType maps payment status to Kafka event type.
func mapStatusToEventType(status constant.PaymentStatus) string {
	switch status {
	case constant.PaymentStatusPending:
		return event.PaymentCreatedEventType
	case constant.PaymentStatusProcessing:
		return event.PaymentProcessingEventType
	case constant.PaymentStatusCompleted:
		return event.PaymentCompletedEventType
	case constant.PaymentStatusFailed:
		return event.PaymentFailedEventType
	case constant.PaymentStatusRefunded:
		return event.PaymentRefundedEventType
	default:
		return "unknown"
	}
}
