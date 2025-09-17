package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// ProductDeletedEvent is the envelope for product deletion events.
type ProductDeletedEvent struct {
	Metadata event.Metadata              `json:"metadata"`
	Payload  event.ProductDeletedPayload `json:"payload"`
}

// NewProductDeletedEvent creates a new ProductDeletedEvent.
func NewProductDeletedEvent(productID uuid.UUID) *ProductDeletedEvent {
	return &ProductDeletedEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.ProductDeletedEventType,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.ProductServiceName,
		},
		Payload: event.ProductDeletedPayload{
			ProductID: productID,
		},
	}
}

// GetPayload returns the data associated with the ProductDeletedEvent.
func (e *ProductDeletedEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the ProductDeletedEvent.
func (e *ProductDeletedEvent) GetMetadata() event.Metadata {
	return e.Metadata
}

// ProductDeletedProducer is responsible for producing product deleted events.
type ProductDeletedProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewProductDeletedProducer creates a new instance of ProductDeletedProducer.
func NewProductDeletedProducer(producer *kafka.AsyncProducer) kafka.Producer {
	return &ProductDeletedProducer{
		Producer: producer,
		topic:    kafka.ProductLifecycleTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *ProductDeletedProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *ProductDeletedProducer) Topic() string {
	return p.topic
}
