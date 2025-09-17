// Package producer contains the Kafka producer for Fulfillment DLQ events.
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

// FulfillmentDLQEvent is the envelope for all Fulfillment DLQ events.
type FulfillmentDLQEvent struct {
	Metadata event.Metadata   `json:"metadata"`
	Payload  event.DLQPayload `json:"payload"`
}

// FulfillmentDLQProducer is responsible for producing Fulfillment DLQ events.
type FulfillmentDLQProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewFulfillmentDLQEvent creates a new FulfillmentDLQEvent.
func NewFulfillmentDLQEvent(
	outboxEvent *entity.OutboxEvent,
	reason pkgconstant.DLQReason,
) *FulfillmentDLQEvent {
	return &FulfillmentDLQEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.FulfillmentDLQEventType,
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

// GetPayload returns the data associated with the FulfillmentDLQEvent.
func (e *FulfillmentDLQEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the FulfillmentDLQEvent.
func (e *FulfillmentDLQEvent) GetMetadata() event.Metadata {
	return e.Metadata
}

// NewFulfillmentDLQProducer creates a new instance of FulfillmentDLQProducer.
func NewFulfillmentDLQProducer(producer *kafka.AsyncProducer) kafka.Producer {
	return &FulfillmentDLQProducer{
		Producer: producer,
		topic:    kafka.FulfillmentDLQTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *FulfillmentDLQProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *FulfillmentDLQProducer) Topic() string {
	return p.topic
}
