package payload

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// OrderItemPayload holds the data for each item in the order.
type OrderItemPayload struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int64     `json:"quantity"`
}

// OrderLifecyclePayload holds the data for the Order Lifecycle event.
type OrderLifecyclePayload struct {
	OrderID    uuid.UUID          `json:"order_id"`
	UserID     uuid.UUID          `json:"user_id"`
	Status     string             `json:"status"`
	TotalPrice decimal.Decimal    `json:"total_price"`
	Items      []OrderItemPayload `json:"items"`
}
