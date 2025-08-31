package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/mq"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
)

// ProductCreatedPayload holds the data for the product created event.
type ProductCreatedPayload struct {
	ProductID uuid.UUID       `json:"product_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Quantity  int64           `json:"quantity"`
}

// ProductCreatedEvent is the envelope for all product events.
type ProductCreatedEvent struct {
	Metadata mq.KafkaMetadata      `json:"metadata"`
	Payload  ProductCreatedPayload `json:"payload"`
}

// GetPayload returns the data associated with the ProductCreatedEvent.
func (e *ProductCreatedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the ProductCreatedEvent.
func (e *ProductCreatedEvent) GetMetadata() mq.KafkaMetadata { // Use the correct type from mq package
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
		Metadata: mq.KafkaMetadata{ // Use the correct type from mq package
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeProductCreated,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceProductService,
		},
		Payload: ProductCreatedPayload{
			ProductID: productID,
			Name:      name,
			Price:     price,
			Quantity:  quantity,
		},
	}
}

// ProductCreatedProducer is responsible for producing product created events.
type ProductCreatedProducer struct {
	Producer *mq.KafkaAsyncProducer
	topic    string
}

// NewProductCreatedProducer creates a new instance of ProductCreatedProducer.
func NewProductCreatedProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducerInterface {
	return &ProductCreatedProducer{
		Producer: producer,
		topic:    constant.TopicProductLifecycle,
	}
}

// Send implements the KafkaProducer interface.
func (p *ProductCreatedProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *ProductCreatedProducer) Topic() string {
	return p.topic
}
