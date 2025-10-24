// Package producer contains the Kafka producer for Payment events.
package producer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"
	"github.com/shopspring/decimal"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/mapper"
)

// PaymentLifecycleEvent is the envelope for all Payment events.
type PaymentLifecycleEvent struct {
	Metadata kafkaevent.Metadata                `json:"metadata"`
	Payload  kafkaevent.PaymentLifecyclePayload `json:"payload"`
}

// PaymentLifecycleProducer is responsible for producing Payment Lifecycle events.
type PaymentLifecycleProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewPaymentLifecycleEvent creates a new PaymentLifecycleEvent.
func NewPaymentLifecycleEvent(
	paymentID uuid.UUID,
	orderID uuid.UUID,
	newStatus constant.PaymentStatus,
	totalPrice decimal.Decimal,
	clientSecret *string,
	expiresAt *time.Time,
) *PaymentLifecycleEvent {
	return &PaymentLifecycleEvent{
		Metadata: kafkaevent.Metadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   mapper.MapStatusToEventType(newStatus),
			AggregateID: paymentID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.PaymentServiceName,
		},
		Payload: kafkaevent.PaymentLifecyclePayload{
			PaymentID:    paymentID,
			OrderID:      orderID,
			Status:       string(newStatus),
			TotalPrice:   totalPrice,
			ClientSecret: clientSecret, // Stripe client secret for Payment Element
			ExpiresAt:    expiresAt,    // 24-hour payment window expiry
		},
	}
}

// GetPayload returns the data associated with the PaymentLifecycleEvent.
func (e *PaymentLifecycleEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the PaymentLifecycleEvent.
func (e *PaymentLifecycleEvent) GetMetadata() kafkaevent.Metadata { // Use the correct type from mq package
	return e.Metadata
}

// NewPaymentLifecycleProducer creates a new instance of PaymentLifecycleProducer.
func NewPaymentLifecycleProducer(producer *kafka.AsyncProducer) kafka.Producer {
	return &PaymentLifecycleProducer{
		Producer: producer,
		topic:    kafka.PaymentLifecycleTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *PaymentLifecycleProducer) Send(ctx context.Context, evt kafkaevent.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *PaymentLifecycleProducer) Topic() string {
	return p.topic
}
