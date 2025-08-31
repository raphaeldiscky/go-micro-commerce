package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
)

// PaymentDLQPayload holds the data for the Payment DLQ event.
type PaymentDLQPayload struct {
	OutboxEventID   uuid.UUID          `json:"outbox_event_id"`
	AggregateType   string             `json:"aggregate_type"`
	AggregateID     uuid.UUID          `json:"aggregate_id"`
	OriginalTopic   string             `json:"original_topic"`
	OriginalPayload json.RawMessage    `json:"original_payload"`
	Reason          constant.DLQReason `json:"reason"`
	LastError       string             `json:"last_error"`
	Attempts        int64              `json:"attempts"`
	CreatedAt       time.Time          `json:"created_at"`
	LastProcessedAt *time.Time         `json:"last_processed_at"`
	FailedAt        time.Time          `json:"failed_at"`
}

// PaymentDLQEvent is the envelope for all Payment events.
type PaymentDLQEvent struct {
	Metadata mq.KafkaMetadata  `json:"metadata"`
	Payload  PaymentDLQPayload `json:"payload"`
}

// GetPayload returns the data associated with the PaymentDLQEvent.
func (e *PaymentDLQEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the PaymentDLQEvent.
func (e *PaymentDLQEvent) GetMetadata() mq.KafkaMetadata { // Use the correct type from mq package
	return e.Metadata
}

// PaymentDLQProducer is responsible for producing Payment DLQ events.
type PaymentDLQProducer struct {
	Producer *mq.KafkaAsyncProducer
	topic    string
}

// NewPaymentDLQEvent creates a new PaymentDLQEvent.
func NewPaymentDLQEvent(
	outboxEvent *entity.OutboxEvent,
	reason constant.DLQReason,
) *PaymentDLQEvent {
	return &PaymentDLQEvent{
		Metadata: mq.KafkaMetadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypePaymentDLQ,
			AggregateID: outboxEvent.AggregateID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourcePaymentService,
		},
		Payload: PaymentDLQPayload{
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

// NewPaymentDLQProducer creates a new instance of PaymentDLQProducer.
func NewPaymentDLQProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducerInterface {
	return &PaymentDLQProducer{
		Producer: producer,
		topic:    constant.TopicPaymentDLQ,
	}
}

// Send implements the KafkaProducer interface.
func (p *PaymentDLQProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *PaymentDLQProducer) Topic() string {
	return p.topic
}
