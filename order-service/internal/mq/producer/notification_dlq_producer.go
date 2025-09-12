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

// NotificationDLQEvent is the envelope for all Fulfillment DLQ events.
type NotificationDLQEvent struct {
	Metadata event.Metadata   `json:"metadata"`
	Payload  event.DLQPayload `json:"payload"`
}

// NotificationDLQProducer is responsible for producing Fulfillment DLQ events.
type NotificationDLQProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewNotificationDLQEvent creates a new NotificationDLQEvent.
func NewNotificationDLQEvent(
	outboxEvent *entity.OutboxEvent,
	reason pkgconstant.DLQReason,
) *NotificationDLQEvent {
	return &NotificationDLQEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.NotificationDLQEventType,
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

// GetPayload returns the data associated with the NotificationDLQEvent.
func (e *NotificationDLQEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the NotificationDLQEvent.
func (e *NotificationDLQEvent) GetMetadata() event.Metadata {
	return e.Metadata
}

// NewNotificationDLQProducer creates a new instance of NotificationDLQProducer.
func NewNotificationDLQProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &NotificationDLQProducer{
		Producer: producer,
		topic:    kafka.NotificationDLQTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *NotificationDLQProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *NotificationDLQProducer) Topic() string {
	return p.topic
}
