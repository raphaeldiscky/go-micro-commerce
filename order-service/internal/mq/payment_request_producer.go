package mq

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

// PaymentGatewayRequestEvent is the envelope for payment request events.
type PaymentGatewayRequestEvent struct {
	Metadata event.Metadata                     `json:"metadata"`
	Payload  event.PaymentGatewayRequestPayload `json:"payload"`
}

// PaymentGatewayRequestProducer is responsible for producing Payment Lifecycle events.
type PaymentGatewayRequestProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// GetPayload returns the data associated with the PaymentGatewayRequestEvent.
func (e *PaymentGatewayRequestEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the PaymentGatewayRequestEvent.
func (e *PaymentGatewayRequestEvent) GetMetadata() event.Metadata {
	return e.Metadata
}

// NewPaymentGatewayRequestEvent creates a new PaymentGatewayRequestEvent.
func NewPaymentGatewayRequestEvent(
	orderID uuid.UUID,
	customerID uuid.UUID,
	totalPrice decimal.Decimal,
	currency string,
	paymentMethod constant.PaymentMethod,
) *PaymentGatewayRequestEvent {
	return &PaymentGatewayRequestEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.PaymentGatewayRequestedEventType,
			AggregateID: orderID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.OrderServiceName,
		},
		Payload: event.PaymentGatewayRequestPayload{
			PaymentID:     uuid.New(),
			OrderID:       orderID,
			CustomerID:    customerID,
			TotalPrice:    totalPrice,
			Currency:      currency,
			PaymentMethod: string(paymentMethod),
		},
	}
}

// NewPaymentGatewayRequestProducer creates a new instance of PaymentGatewayRequestProducer.
func NewPaymentGatewayRequestProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &PaymentGatewayRequestProducer{
		Producer: producer,
		topic:    kafka.PaymentGatewayRequestTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *PaymentGatewayRequestProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *PaymentGatewayRequestProducer) Topic() string {
	return p.topic
}
