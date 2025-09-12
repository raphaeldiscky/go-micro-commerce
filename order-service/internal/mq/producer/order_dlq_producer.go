// Package producer contains the Kafka producer for Order events.
package producer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// OrderDLQEvent is the envelope for all Order events.
type OrderDLQEvent struct {
	Metadata event.Metadata   `json:"metadata"`
	Payload  event.DLQPayload `json:"payload"`
}

// OrderDLQProducer is responsible for producing Order DLQ events.
type OrderDLQProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewOrderDLQEvent creates a new OrderDLQEvent.
func NewOrderDLQEvent(
	outboxEvent *entity.OutboxEvent,
	reason pkgconstant.DLQReason,
) *OrderDLQEvent {
	return &OrderDLQEvent{
		Metadata: event.Metadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   kafka.OrderDLQEventType,
			AggregateID: outboxEvent.AggregateID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.OrderServiceName,
		},
		Payload: event.DLQPayload{
			OutboxEventID:   outboxEvent.ID,
			AggregateType:   outboxEvent.AggregateType,
			AggregateID:     outboxEvent.AggregateID,
			OriginalTopic:   outboxEvent.Topic,
			OriginalPayload: outboxEvent.Payload,
			Reason:          reason,
			LastError:       *outboxEvent.LastError,
			Attempts:        outboxEvent.Attempts,
			CreatedAt:       outboxEvent.CreatedAt,
			LastProcessedAt: outboxEvent.ProcessedAt,
			FailedAt:        time.Now().UTC(),
		},
	}
}

// GetPayload returns the data associated with the OrderDLQEvent.
func (e *OrderDLQEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the OrderDLQEvent.
func (e *OrderDLQEvent) GetMetadata() event.Metadata { // Use the correct type from mq package
	return e.Metadata
}

// NewOrderDLQProducer creates a new instance of OrderDLQProducer.
func NewOrderDLQProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &OrderDLQProducer{
		Producer: producer,
		topic:    kafka.OrderDLQTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *OrderDLQProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *OrderDLQProducer) Topic() string {
	return p.topic
}
