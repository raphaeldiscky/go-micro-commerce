// Package producer contains the Kafka producer for CheckoutSession events.
package producer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// NotificationRequestEvent is the envelope for NotificationRequest events.
type NotificationRequestEvent struct {
	Metadata kafkaevent.Metadata                   `json:"metadata"`
	Payload  kafkaevent.NotificationRequestPayload `json:"payload"`
}

// NotificationRequestProducer is responsible for producing CheckoutSession OrderPlaced events.
type NotificationRequestProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewNotificationRequestEvent creates a new NotificationRequestEvent.
func NewNotificationRequestEvent(
	checkoutSessionID uuid.UUID,
	customerEmail string,
) *NotificationRequestEvent {
	return &NotificationRequestEvent{
		Metadata: kafkaevent.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.NotificationRequestedEventType,
			AggregateID: checkoutSessionID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.CartServiceName,
		},
		Payload: kafkaevent.NotificationRequestPayload{
			NotificationType: kafkaevent.NotificationTypePush,
			RecipientEmail:   customerEmail,
			CreatedAt:        time.Now().UTC(),
		},
	}
}

// GetPayload returns the data associated with the NotificationRequestEvent.
func (e *NotificationRequestEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the NotificationRequestEvent.
func (e *NotificationRequestEvent) GetMetadata() kafkaevent.Metadata {
	return e.Metadata
}

// NewNotificationRequestProducer creates a new instance of NotificationRequestProducer.
func NewNotificationRequestProducer(producer *kafka.AsyncProducer) kafka.Producer {
	return &NotificationRequestProducer{
		Producer: producer,
		topic:    kafka.NotificationRequestTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *NotificationRequestProducer) Send(
	ctx context.Context,
	evt kafkaevent.BaseEvent,
) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *NotificationRequestProducer) Topic() string {
	return p.topic
}
