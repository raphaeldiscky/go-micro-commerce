package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/entity"
)

// FulfillmentDLQEvent is the envelope for all Fulfillment DLQ events.
type FulfillmentDLQEvent struct {
	Metadata kafkaevent.Metadata   `json:"metadata"`
	Payload  kafkaevent.DLQPayload `json:"payload"`
}

// FulfillmentDLQProducer is responsible for producing fulfillment events to the Dead Letter Queue.
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
		Metadata: kafkaevent.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.FulfillmentDLQEventType,
			AggregateID: outboxEvent.AggregateID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.FulfillmentServiceName,
		},
		Payload: kafkaevent.DLQPayload{
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
func (e *FulfillmentDLQEvent) GetMetadata() kafkaevent.Metadata {
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
func (p *FulfillmentDLQProducer) Send(ctx context.Context, evt kafkaevent.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *FulfillmentDLQProducer) Topic() string {
	return p.topic
}
