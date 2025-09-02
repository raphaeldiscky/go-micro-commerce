package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event/payload"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/shopspring/decimal"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// OrderLifecycleEvent is the envelope for all Order events.
type OrderLifecycleEvent struct {
	Metadata event.Metadata                `json:"metadata"`
	Payload  payload.OrderLifecyclePayload `json:"payload"`
}

// GetPayload returns the data associated with the OrderLifecycleEvent.
func (e *OrderLifecycleEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the OrderLifecycleEvent.
func (e *OrderLifecycleEvent) GetMetadata() event.Metadata { // Use the correct type from mq package
	return e.Metadata
}

// OrderLifecycleProducer is responsible for producing Order Lifecycle events.
type OrderLifecycleProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewOrderLifecycleEvent creates a new OrderLifecycleEvent.
func NewOrderLifecycleEvent(
	orderID uuid.UUID,
	newStatus constant.OrderStatus,
	userID uuid.UUID,
	totalPrice decimal.Decimal,
	items []entity.OrderItem,
) *OrderLifecycleEvent {
	return &OrderLifecycleEvent{
		Metadata: event.Metadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   mapStatusToEventType(newStatus),
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.OrderServiceName,
		},
		Payload: payload.OrderLifecyclePayload{
			OrderID:    orderID,
			UserID:     userID,
			Status:     string(newStatus),
			TotalPrice: totalPrice,
			Items:      mapOrderItemsToPayload(items),
		},
	}
}

// NewOrderLifecycleProducer creates a new instance of OrderLifecycleProducer.
func NewOrderLifecycleProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &OrderLifecycleProducer{
		Producer: producer,
		topic:    event.OrderLifecycleTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *OrderLifecycleProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *OrderLifecycleProducer) Topic() string {
	return p.topic
}
