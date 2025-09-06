package event

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PaymentRequestPayload holds the data for payment request events.
type PaymentRequestPayload struct {
	PaymentID     uuid.UUID       `json:"payment_id"`
	OrderID       uuid.UUID       `json:"order_id"`
	CustomerID    uuid.UUID       `json:"customer_id"`
	TotalPrice    decimal.Decimal `json:"total_price"`
	Currency      string          `json:"currency"`
	PaymentMethod string          `json:"payment_method"`
}

// PaymentLifecyclePayload holds the data for the Payment Lifecycle event.
type PaymentLifecyclePayload struct {
	PaymentID  uuid.UUID       `json:"payment_id"`
	OrderID    uuid.UUID       `json:"order_id"`
	Status     string          `json:"status"`
	TotalPrice decimal.Decimal `json:"total_price"`
}
