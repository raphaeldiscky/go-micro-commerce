// Package mq provides the event definitions and handlers for the product service.
package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/shopspring/decimal"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// ProductCreatedEvent is the envelope for all product events.
type ProductCreatedEvent struct {
	Metadata event.Metadata              `json:"metadata"`
	Payload  event.ProductCreatedPayload `json:"payload"`
}

// NewProductCreatedEvent creates a new ProductCreatedEvent.
func NewProductCreatedEvent(
	productID uuid.UUID,
	name string,
	price decimal.Decimal,
	quantity int64,
) *ProductCreatedEvent {
	return &ProductCreatedEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.ProductCreatedEventType,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.ProductServiceName,
		},
		Payload: event.ProductCreatedPayload{
			ProductID: productID,
			Name:      name,
			Price:     price,
			Quantity:  quantity,
		},
	}
}

// GetPayload returns the data associated with the ProductCreatedEvent.
func (e *ProductCreatedEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the ProductCreatedEvent.
func (e *ProductCreatedEvent) GetMetadata() event.Metadata { // Use the correct type from mq package
	return e.Metadata
}

// ProductCreatedProducer is responsible for producing product created events.
type ProductCreatedProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewProductCreatedProducer creates a new instance of ProductCreatedProducer.
func NewProductCreatedProducer(producer *kafka.AsyncProducer) kafka.Producer {
	return &ProductCreatedProducer{
		Producer: producer,
		topic:    kafka.ProductLifecycleTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *ProductCreatedProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *ProductCreatedProducer) Topic() string {
	return p.topic
}
