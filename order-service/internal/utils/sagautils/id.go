// Package sagautils provides utility functions for working with saga IDs.
package sagautils

import (
	"fmt"

	"github.com/google/uuid"
)

// CreateOrderSagaID creates a unique ID for the order saga.
func CreateOrderSagaID(orderID uuid.UUID) string {
	return fmt.Sprintf("order:saga:%s", orderID)
}

// CreatePaymentReminderID creates a unique ID for the payment reminder.
func CreatePaymentReminderID(orderID uuid.UUID) string {
	return fmt.Sprintf("payment:reminder:%s", orderID)
}

// GenerateTaskID creates a unique task ID for asynq tasks with correlation metadata.
func GenerateTaskID(orderID, correlationID uuid.UUID, reminderCount int) string {
	return fmt.Sprintf(
		"order:%s:corr:%s:reminder:%d",
		orderID.String(),
		correlationID.String(),
		reminderCount,
	)
}

// GenerateCancelTaskID creates a unique task ID for order cancellation tasks.
func GenerateCancelTaskID(orderID, correlationID uuid.UUID) string {
	return fmt.Sprintf("order:%s:corr:%s:cancel", orderID.String(), correlationID.String())
}

// GenerateTaskPattern creates a pattern to match all tasks for a specific order and correlation.
func GenerateTaskPattern(orderID, correlationID uuid.UUID) string {
	return fmt.Sprintf("order:%s:corr:%s", orderID.String(), correlationID.String())
}
