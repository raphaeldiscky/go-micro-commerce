package producer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
)

// PaymentDLQEvent is the envelope for all Payment events.
type PaymentDLQEvent struct {
	Payload  kafkaevent.DLQPayload `json:"payload"`
	Metadata kafkaevent.Metadata   `json:"metadata"`
}

// PaymentDLQProducer is responsible for producing Payment DLQ events.
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
		Metadata: kafkaevent.Metadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   kafka.PaymentDLQEventType,
			AggregateID: outboxEvent.AggregateID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.PaymentServiceName,
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

// GetPayload returns the data associated with the PaymentDLQEvent.
func (e *PaymentDLQEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the PaymentDLQEvent.
func (e *PaymentDLQEvent) GetMetadata() kafkaevent.Metadata { // Use the correct type from mq package
	return e.Metadata
}

// NewPaymentDLQProducer creates a new instance of PaymentDLQProducer.
func NewPaymentDLQProducer(producer *kafka.AsyncProducer) kafka.Producer {
	return &PaymentDLQProducer{
		Producer: producer,
		topic:    kafka.PaymentDLQTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *PaymentDLQProducer) Send(ctx context.Context, evt kafkaevent.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *PaymentDLQProducer) Topic() string {
	return p.topic
}
