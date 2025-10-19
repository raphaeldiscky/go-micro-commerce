package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"
	"github.com/shopspring/decimal"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// ProductUpdatedEvent is the envelope for product update events.
type ProductUpdatedEvent struct {
	Metadata kafkaevent.Metadata              `json:"metadata"`
	Payload  kafkaevent.ProductUpdatedPayload `json:"payload"`
}

// NewProductUpdatedEvent creates a new ProductUpdatedEvent.
func NewProductUpdatedEvent(
	productID uuid.UUID,
	name string,
	price decimal.Decimal,
	quantity int64,
) *ProductUpdatedEvent {
	return &ProductUpdatedEvent{
		Metadata: kafkaevent.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.ProductUpdatedEventType,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.ProductServiceName,
		},
		Payload: kafkaevent.ProductUpdatedPayload{
			ProductID: productID,
			Name:      name,
			Price:     price,
			Quantity:  quantity,
		},
	}
}

// GetPayload returns the data associated with the ProductUpdatedEvent.
func (e *ProductUpdatedEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the ProductUpdatedEvent.
func (e *ProductUpdatedEvent) GetMetadata() kafkaevent.Metadata {
	return e.Metadata
}

// ProductUpdatedProducer is responsible for producing product Updated events.
type ProductUpdatedProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewProductUpdatedProducer creates a new instance of ProductUpdatedProducer.
func NewProductUpdatedProducer(producer *kafka.AsyncProducer) kafka.Producer {
	return &ProductUpdatedProducer{
		Producer: producer,
		topic:    kafka.ProductLifecycleTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *ProductUpdatedProducer) Send(ctx context.Context, evt kafkaevent.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *ProductUpdatedProducer) Topic() string {
	return p.topic
}
