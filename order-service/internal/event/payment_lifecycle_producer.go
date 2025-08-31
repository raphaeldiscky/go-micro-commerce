package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// PaymentLifecyclePayload holds the data for payment request events.
type PaymentLifecyclePayload struct {
	OrderID       uuid.UUID              `json:"order_id"`
	CustomerID    uuid.UUID              `json:"customer_id"`
	TotalPrice    decimal.Decimal        `json:"total_price"`
	Currency      string                 `json:"currency"`
	PaymentMethod constant.PaymentMethod `json:"payment_method"`
}

// PaymentLifecycleEvent is the envelope for payment request events.
type PaymentLifecycleEvent struct {
	Metadata mq.KafkaMetadata        `json:"metadata"`
	Payload  PaymentLifecyclePayload `json:"payload"`
}

// PaymentLifecycleProducer is responsible for producing Payment Lifecycle events.
type PaymentLifecycleProducer struct {
	Producer *mq.KafkaAsyncProducer
	topic    string
}

// GetPayload returns the data associated with the PaymentLifecycleEvent.
func (e *PaymentLifecycleEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the PaymentLifecycleEvent.
func (e *PaymentLifecycleEvent) GetMetadata() mq.KafkaMetadata {
	return e.Metadata
}

// NewPaymentLifecycleEvent creates a new PaymentLifecycleEvent.
func NewPaymentLifecycleEvent(
	orderID uuid.UUID,
	customerID uuid.UUID,
	totalPrice decimal.Decimal,
	currency string,
	paymentMethod constant.PaymentMethod,
) *PaymentLifecycleEvent {
	return &PaymentLifecycleEvent{
		Metadata: mq.KafkaMetadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypePaymentRequested,
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceOrderService,
		},
		Payload: PaymentLifecyclePayload{
			OrderID:       orderID,
			CustomerID:    customerID,
			TotalPrice:    totalPrice,
			Currency:      currency,
			PaymentMethod: paymentMethod,
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
