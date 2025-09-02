package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event/payload"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
)

// ProductCreatedEvent is the envelope for all product events.
type ProductCreatedEvent struct {
	Metadata event.Metadata                `json:"metadata"`
	Payload  payload.ProductCreatedPayload `json:"payload"`
}

// GetPayload returns the data associated with the ProductCreatedEvent.
func (e *ProductCreatedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the ProductCreatedEvent.
func (e *ProductCreatedEvent) GetMetadata() event.Metadata { // Use the correct type from mq package
	return e.Metadata
}

// NewProductCreatedEvent creates a new ProductCreatedEvent.
func NewProductCreatedEvent(
	productID uuid.UUID,
	name string,
	price decimal.Decimal,
	quantity int64,
) *ProductCreatedEvent {
	return &ProductCreatedEvent{
		Metadata: event.Metadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeProductCreated,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceProductService,
		},
		Payload: payload.ProductCreatedPayload{
			ProductID: productID,
			Name:      name,
			Price:     price,
			Quantity:  quantity,
		},
	}
}

// ProductCreatedProducer is responsible for producing product created events.
type ProductCreatedProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewProductCreatedProducer creates a new instance of ProductCreatedProducer.
func NewProductCreatedProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &ProductCreatedProducer{
		Producer: producer,
		topic:    constant.TopicProductLifecycle,
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
