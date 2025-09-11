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

// PaymentDLQEvent is the envelope for all Order events.
type PaymentDLQEvent struct {
	Metadata event.Metadata   `json:"metadata"`
	Payload  event.DLQPayload `json:"payload"`
}

// PaymentDLQProducer is responsible for producing Order DLQ events.
type PaymentDLQProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewPaymentDLQEvent creates a new PaymentDLQEvent.
func NewPaymentDLQEvent(
	outboxEvent *entity.OutboxEvent,
	reason pkgconstant.DLQReason,
) *PaymentDLQEvent {
	return &PaymentDLQEvent{
		Metadata: event.Metadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   kafka.PaymentDLQEventType,
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

// GetPayload returns the data associated with the PaymentDLQEvent.
func (e *PaymentDLQEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the PaymentDLQEvent.
func (e *PaymentDLQEvent) GetMetadata() event.Metadata { // Use the correct type from mq package
	return e.Metadata
}

// NewPaymentDLQProducer creates a new instance of PaymentDLQProducer.
func NewPaymentDLQProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &PaymentDLQProducer{
		Producer: producer,
		topic:    kafka.PaymentDLQTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *PaymentDLQProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *PaymentDLQProducer) Topic() string {
	return p.topic
}
