package producer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/shopspring/decimal"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// PaymentRequestEvent is the envelope for payment request events.
type PaymentRequestEvent struct {
	Metadata event.Metadata              `json:"metadata"`
	Payload  event.PaymentRequestPayload `json:"payload"`
}

// PaymentRequestProducer is responsible for producing Payment Lifecycle events.
type PaymentRequestProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
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
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.PaymentRequestedEventType,
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.OrderServiceName,
		},
		Payload: event.PaymentRequestPayload{
			PaymentID:     uuid.New(),
			OrderID:       orderID,
			CustomerID:    customerID,
			TotalPrice:    totalPrice,
			Currency:      currency,
			PaymentMethod: string(paymentMethod),
		},
	}
}

// GetPayload returns the data associated with the PaymentRequestEvent.
func (e *PaymentRequestEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the PaymentRequestEvent.
func (e *PaymentRequestEvent) GetMetadata() event.Metadata {
	return e.Metadata
}

// NewPaymentRequestProducer creates a new instance of PaymentRequestProducer.
func NewPaymentRequestProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &PaymentRequestProducer{
		Producer: producer,
		topic:    kafka.PaymentRequestTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *PaymentRequestProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *PaymentRequestProducer) Topic() string {
	return p.topic
}
