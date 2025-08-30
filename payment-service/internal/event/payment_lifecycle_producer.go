package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// PaymentLifecyclePayload holds the data for the Payment Lifecycle event.
type PaymentLifecyclePayload struct {
	PaymentID  uuid.UUID              `json:"order_id"`
	Status     constant.PaymentStatus `json:"status"`
	TotalPrice decimal.Decimal        `json:"total_price"`
}

// PaymentLifecycleEvent is the envelope for all Payment events.
type PaymentLifecycleEvent struct {
	Metadata mq.KafkaMetadata        `json:"metadata"`
	Payload  PaymentLifecyclePayload `json:"payload"`
}

// GetPayload returns the data associated with the PaymentLifecycleEvent.
func (e *PaymentLifecycleEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the PaymentLifecycleEvent.
func (e *PaymentLifecycleEvent) GetMetadata() mq.KafkaMetadata { // Use the correct type from mq package
	return e.Metadata
}

// PaymentLifecycleProducer is responsible for producing Payment Lifecycle events.
type PaymentLifecycleProducer struct {
	Producer *mq.KafkaAsyncProducer
	topic    string
}

// NewPaymentLifecycleEvent creates a new PaymentLifecycleEvent.
func NewPaymentLifecycleEvent(
	orderID uuid.UUID,
	newStatus constant.PaymentStatus,
	totalPrice decimal.Decimal,
) *PaymentLifecycleEvent {
	return &PaymentLifecycleEvent{
		Metadata: mq.KafkaMetadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   mapStatusToEventType(newStatus),
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourcePaymentService,
		},
		Payload: PaymentLifecyclePayload{
			PaymentID:  orderID,
			Status:     newStatus,
			TotalPrice: totalPrice,
		},
	}
}

// NewPaymentLifecycleProducer creates a new instance of PaymentLifecycleProducer.
func NewPaymentLifecycleProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducerInterface {
	return &PaymentLifecycleProducer{
		Producer: producer,
		topic:    constant.TopicPaymentLifecycle,
	}
}

// Send implements the KafkaProducer interface.
func (p *PaymentLifecycleProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *PaymentLifecycleProducer) Topic() string {
	return p.topic
}
