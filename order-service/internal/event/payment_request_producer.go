package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// PaymentRequestPayload holds the data for payment request events.
type PaymentRequestPayload struct {
	PaymentID     uuid.UUID              `json:"payment_id"`
	OrderID       uuid.UUID              `json:"order_id"`
	CustomerID    uuid.UUID              `json:"customer_id"`
	TotalPrice    decimal.Decimal        `json:"total_price"`
	Currency      string                 `json:"currency"`
	PaymentMethod constant.PaymentMethod `json:"payment_method"`
}

// PaymentRequestEvent is the envelope for payment request events.
type PaymentRequestEvent struct {
	Metadata kafka.Metadata        `json:"metadata"`
	Payload  PaymentRequestPayload `json:"payload"`
}

// PaymentRequestProducer is responsible for producing Payment Lifecycle events.
type PaymentRequestProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// GetPayload returns the data associated with the PaymentRequestEvent.
func (e *PaymentRequestEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the PaymentRequestEvent.
func (e *PaymentRequestEvent) GetMetadata() kafka.Metadata {
	return e.Metadata
}

// NewPaymentRequestEvent creates a new PaymentRequestEvent.
func NewPaymentRequestEvent(
	orderID uuid.UUID,
	customerID uuid.UUID,
	totalPrice decimal.Decimal,
	currency string,
	paymentMethod constant.PaymentMethod,
) *PaymentRequestEvent {
	return &PaymentRequestEvent{
		Metadata: kafka.Metadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypePaymentRequested,
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceOrderService,
		},
		Payload: PaymentRequestPayload{
			PaymentID:     uuid.New(),
			OrderID:       orderID,
			CustomerID:    customerID,
			TotalPrice:    totalPrice,
			Currency:      currency,
			PaymentMethod: paymentMethod,
		},
	}
}

// NewPaymentRequestProducer creates a new instance of PaymentRequestProducer.
func NewPaymentRequestProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &PaymentRequestProducer{
		Producer: producer,
		topic:    constant.TopicPaymentRequest,
	}
}

// Send implements the KafkaProducer interface.
func (p *PaymentRequestProducer) Send(ctx context.Context, event kafka.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *PaymentRequestProducer) Topic() string {
	return p.topic
}
