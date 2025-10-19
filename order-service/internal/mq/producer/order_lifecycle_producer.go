package producer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"
	"github.com/shopspring/decimal"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/mapper"
)

// OrderLifecycleEvent is the envelope for all Order events.
type OrderLifecycleEvent struct {
	Metadata kafkaevent.Metadata              `json:"metadata"`
	Payload  kafkaevent.OrderLifecyclePayload `json:"payload"`
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
		Metadata: kafkaevent.Metadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   mapper.MapStatusToEventType(newStatus),
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.OrderServiceName,
		},
		Payload: kafkaevent.OrderLifecyclePayload{
			OrderID:    orderID,
			UserID:     userID,
			Status:     string(newStatus),
			TotalPrice: totalPrice,
			Items:      mapper.MapOrderItemsToPayload(items),
		},
	}
}

// GetPayload returns the data associated with the OrderLifecycleEvent.
func (e *OrderLifecycleEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the OrderLifecycleEvent.
func (e *OrderLifecycleEvent) GetMetadata() kafkaevent.Metadata { // Use the correct type from mq package
	return e.Metadata
}

// NewOrderLifecycleProducer creates a new instance of OrderLifecycleProducer.
func NewOrderLifecycleProducer(producer *kafka.AsyncProducer) kafka.Producer {
	return &OrderLifecycleProducer{
		Producer: producer,
		topic:    kafka.OrderLifecycleTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *OrderLifecycleProducer) Send(ctx context.Context, evt kafkaevent.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *OrderLifecycleProducer) Topic() string {
	return p.topic
}
