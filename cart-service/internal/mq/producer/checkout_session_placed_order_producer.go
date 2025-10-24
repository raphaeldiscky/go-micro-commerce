// Package producer contains the Kafka producer for CheckoutSession events.
package producer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/mapper"
)

// CheckoutSessionOrderPlacedEvent is the envelope for CheckoutSessionOrderPlaced events.
type CheckoutSessionOrderPlacedEvent struct {
	Metadata kafkaevent.Metadata                          `json:"metadata"`
	Payload  kafkaevent.CheckoutSessionOrderPlacedPayload `json:"payload"`
}

// CheckoutSessionOrderPlacedProducer is responsible for producing CheckoutSession OrderPlaced events.
type CheckoutSessionOrderPlacedProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewCheckoutSessionOrderPlacedEvent creates a new CheckoutSessionOrderPlacedEvent.
func NewCheckoutSessionOrderPlacedEvent(
	checkoutSessionID uuid.UUID,
	idempotencyKey uuid.UUID,
	newStatus constant.CheckoutSessionStatus,
	userID uuid.UUID,
	currency string,
	paymentGateway string,
	items []entity.CheckoutSessionItem,
	createdAt time.Time,
) *CheckoutSessionOrderPlacedEvent {
	return &CheckoutSessionOrderPlacedEvent{
		Metadata: kafkaevent.Metadata{
			EventID:     uuid.New(),
			EventType:   mapper.MapCheckoutSessionStatusToEventType(newStatus),
			AggregateID: checkoutSessionID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.CartServiceName,
		},
		Payload: kafkaevent.CheckoutSessionOrderPlacedPayload{
			CheckoutSessionID: checkoutSessionID,
			IdempotencyKey:    idempotencyKey,
			UserID:            userID,
			Status:            string(newStatus),
			Currency:          currency,
			PaymentGateway:    paymentGateway,
			Items:             mapper.MapCheckoutSessionItemsToPayload(items),
			CreatedAt:         createdAt,
		},
	}
}

// GetPayload returns the data associated with the CheckoutSessionOrderPlacedEvent.
func (e *CheckoutSessionOrderPlacedEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the CheckoutSessionOrderPlacedEvent.
func (e *CheckoutSessionOrderPlacedEvent) GetMetadata() kafkaevent.Metadata {
	return e.Metadata
}

// NewCheckoutSessionOrderPlacedProducer creates a new instance of CheckoutSessionOrderPlacedProducer.
func NewCheckoutSessionOrderPlacedProducer(producer *kafka.AsyncProducer) kafka.Producer {
	return &CheckoutSessionOrderPlacedProducer{
		Producer: producer,
		topic:    kafka.CheckoutSessionLifecycleTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *CheckoutSessionOrderPlacedProducer) Send(
	ctx context.Context,
	evt kafkaevent.BaseEvent,
) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *CheckoutSessionOrderPlacedProducer) Topic() string {
	return p.topic
}
