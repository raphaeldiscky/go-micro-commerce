package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/shopspring/decimal"
)

// OrderLifecyclePayload holds the data for the Order Lifecycle event.
type OrderLifecyclePayload struct {
	OrderID    uuid.UUID            `json:"order_id"`
	UserID     uuid.UUID            `json:"user_id"`
	Status     constant.OrderStatus `json:"status"`
	TotalPrice decimal.Decimal      `json:"total_price"`
	Email      string               `json:"email"`
}

// OrderLifecycleEvent is the envelope for all Order events.
type OrderLifecycleEvent struct {
	Metadata mq.KafkaMetadata      `json:"metadata"`
	Payload  OrderLifecyclePayload `json:"payload"`
}

// GetPayload returns the data associated with the OrderLifecycleEvent.
func (e *OrderLifecycleEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the OrderLifecycleEvent.
func (e *OrderLifecycleEvent) GetMetadata() mq.KafkaMetadata { // Use the correct type from mq package
	return e.Metadata
}

// OrderLifecycleProducer is responsible for producing Order Lifecycle events.
type OrderLifecycleProducer struct {
	Producer *mq.KafkaAsyncProducer
	topic    string
}

// NewOrderLifecycleEvent creates a new OrderLifecycleEvent.
func NewOrderLifecycleEvent(
	orderID uuid.UUID,
	newStatus constant.OrderStatus,
	userID uuid.UUID,
	totalPrice decimal.Decimal,
	email string,
) *OrderLifecycleEvent {
	return &OrderLifecycleEvent{
		Metadata: mq.KafkaMetadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   mapStatusToEventType(newStatus),
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceOrderService,
		},
		Payload: OrderLifecyclePayload{
			OrderID:    orderID,
			UserID:     userID,
			Status:     newStatus,
			TotalPrice: totalPrice,
			Email:      email,
		},
	}
}

// NewOrderLifecycleProducer creates a new instance of OrderLifecycleProducer.
func NewOrderLifecycleProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducerInterface {
	return &OrderLifecycleProducer{
		Producer: producer,
		topic:    constant.TopicOrderLifecycle,
	}
}

// Send implements the KafkaProducer interface.
func (p *OrderLifecycleProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *OrderLifecycleProducer) Topic() string {
	return p.topic
}
