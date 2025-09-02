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

// ProductUpdatedEvent is the envelope for product update events.
type ProductUpdatedEvent struct {
	Metadata event.Metadata                `json:"metadata"`
	Payload  payload.ProductUpdatedPayload `json:"payload"`
}

// GetPayload returns the data associated with the ProductUpdatedEvent.
func (e *ProductUpdatedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the ProductUpdatedEvent.
func (e *ProductUpdatedEvent) GetMetadata() event.Metadata {
	return e.Metadata
}

// NewProductUpdatedEvent creates a new ProductUpdatedEvent.
func NewProductUpdatedEvent(
	productID uuid.UUID,
	name string,
	price decimal.Decimal,
	quantity int64,
) *ProductUpdatedEvent {
	return &ProductUpdatedEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeProductUpdated,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceProductService,
		},
		Payload: payload.ProductUpdatedPayload{
			ProductID: productID,
			Name:      name,
			Price:     price,
			Quantity:  quantity,
		},
	}
}

// ProductUpdatedProducer is responsible for producing product Updated events.
type ProductUpdatedProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewProductUpdatedProducer creates a new instance of ProductUpdatedProducer.
func NewProductUpdatedProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &ProductUpdatedProducer{
		Producer: producer,
		topic:    constant.TopicProductLifecycle,
	}
}

// Send implements the KafkaProducer interface.
func (p *ProductUpdatedProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *ProductUpdatedProducer) Topic() string {
	return p.topic
}
