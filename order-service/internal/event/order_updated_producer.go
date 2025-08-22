package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/shopspring/decimal"
)

// OrderUpdatedPayload holds the data for the Order Updated event.
type OrderUpdatedPayload struct {
	OrderID    uuid.UUID            `json:"order_id"`
	CustomerID uuid.UUID            `json:"customer_id"`
	Status     constant.OrderStatus `json:"status"`
	TotalPrice decimal.Decimal      `json:"total_price"`
	Items      []entity.OrderItem   `json:"items"`
}

// OrderUpdatedEvent is the envelope for all Order events.
type OrderUpdatedEvent struct {
	Metadata mq.KafkaMetadata    `json:"metadata"`
	Payload  OrderUpdatedPayload `json:"payload"`
}

// NewOrderUpdatedEvent creates a new OrderUpdatedEvent.
func NewOrderUpdatedEvent(
	orderID uuid.UUID,
	status constant.OrderStatus,
	customerID uuid.UUID,
	totalPrice decimal.Decimal,
	items []entity.OrderItem,
) *OrderUpdatedEvent {
	return &OrderUpdatedEvent{
		Metadata: mq.KafkaMetadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeOrderUpdated,
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceOrderService,
		},
		Payload: OrderUpdatedPayload{
			OrderID:    orderID,
			CustomerID: customerID,
			Status:     status,
			TotalPrice: totalPrice,
			Items:      items,
		},
	}
}
