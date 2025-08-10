package event

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
)

// ProductCreatedPayload holds the data for the product created event.
type ProductCreatedPayload struct {
	ProductID uuid.UUID       `json:"product_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Quantity  int             `json:"quantity"`
}

// ProductCreatedEvent is the envelope for all product events.
type ProductCreatedEvent struct {
	Metadata mq.KafkaMetadata // Use the correct type from mq package
	Payload  ProductCreatedPayload
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
	quantity int,
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
	Producer  *mq.KafkaAsyncProducer
	RetryChan chan *sarama.ProducerMessage
	topic     string
}

// NewProductCreatedProducer creates a new instance of ProductCreatedProducer.
func NewProductCreatedProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducer {
	pr := &ProductCreatedProducer{
		Producer:  producer,
		topic:     constant.ProductLifecycleTopic,
		RetryChan: make(chan *sarama.ProducerMessage, 100),
	}

	return pr
}

// Send implements the KafkaProducer interface
func (p *ProductCreatedProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(p.topic, event)
}

// Topic returns the topic name
func (p *ProductCreatedProducer) Topic() string {
	return p.topic
}
