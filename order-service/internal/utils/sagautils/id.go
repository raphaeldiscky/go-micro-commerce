// Package sagautils provides utility functions for working with saga IDs.
package sagautils

import (
	"fmt"

	"github.com/google/uuid"
)

// CreateOrderSagaID creates a unique ID for the order saga.
func CreateOrderSagaID(orderID uuid.UUID) string {
	return fmt.Sprintf("order-saga-%s", orderID)
}

// CreatePaymentReminderID creates a unique ID for the payment reminder.
func CreatePaymentReminderID(orderID uuid.UUID) string {
	return fmt.Sprintf("payment-reminder-%s", orderID)
}
