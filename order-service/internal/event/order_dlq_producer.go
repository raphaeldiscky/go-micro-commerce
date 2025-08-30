package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// OrderDLQPayload holds the data for the Order DLQ event.
type OrderDLQPayload struct {
	OutboxEventID   uuid.UUID          `json:"outbox_event_id"`
	AggregateType   string             `json:"aggregate_type"`
	AggregateID     uuid.UUID          `json:"aggregate_id"`
	OriginalTopic   string             `json:"original_topic"`
	OriginalPayload json.RawMessage    `json:"original_payload"`
	Reason          constant.DLQReason `json:"reason"`
	LastError       string             `json:"last_error"`
	Attempts        int                `json:"attempts"`
	CreatedAt       time.Time          `json:"created_at"`
	LastProcessedAt *time.Time         `json:"last_processed_at"`
	FailedAt        time.Time          `json:"failed_at"`
}

// OrderDLQEvent is the envelope for all Order events.
type OrderDLQEvent struct {
	Metadata mq.KafkaMetadata `json:"metadata"`
	Payload  OrderDLQPayload  `json:"payload"`
}

// GetPayload returns the data associated with the OrderDLQEvent.
func (e *OrderDLQEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the OrderDLQEvent.
func (e *OrderDLQEvent) GetMetadata() mq.KafkaMetadata { // Use the correct type from mq package
	return e.Metadata
}

// OrderDLQProducer is responsible for producing Order DLQ events.
type OrderDLQProducer struct {
	Producer *mq.KafkaAsyncProducer
	topic    string
}

// NewOrderDLQEvent creates a new OrderDLQEvent.
func NewOrderDLQEvent(
	outboxEvent *entity.OutboxEvent,
	reason constant.DLQReason,
) *OrderDLQEvent {
	return &OrderDLQEvent{
		Metadata: mq.KafkaMetadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeOrderDLQ,
			AggregateID: outboxEvent.AggregateID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceOrderService,
		},
		Payload: OrderDLQPayload{
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

// NewOrderDLQProducer creates a new instance of OrderDLQProducer.
func NewOrderDLQProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducerInterface {
	return &OrderDLQProducer{
		Producer: producer,
		topic:    constant.TopicOrderDLQ,
	}
}

// Send implements the KafkaProducer interface.
func (p *OrderDLQProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *OrderDLQProducer) Topic() string {
	return p.topic
}
