package event

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
)

// ProductUpdatedPayload holds the data for the product updated event.
type ProductUpdatedPayload struct {
	ProductID    uuid.UUID             `json:"product_id"`
	Name         string                `json:"name"`
	Price        decimal.Decimal       `json:"price"`
	Quantity     int                   `json:"quantity"`
	PreviousData ProductCreatedPayload `json:"previous_data,omitempty"` // Optional field for previous product data
}

// ProductUpdatedEvent is the envelope for product update events.
type ProductUpdatedEvent struct {
	Metadata KafkaMetadata
	Payload  ProductUpdatedPayload
}

// GetPayload returns the data associated with the ProductUpdatedEvent.
func (e *ProductUpdatedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the ProductUpdatedEvent.
func (e *ProductUpdatedEvent) GetMetadata() KafkaMetadata {
	return e.Metadata
}

// NewProductUpdatedEvent creates a new ProductUpdatedEvent.
func NewProductUpdatedEvent(
	productID uuid.UUID,
	name string,
	price decimal.Decimal,
	quantity int,
) *ProductUpdatedEvent {
	return &ProductUpdatedEvent{
		Metadata: KafkaMetadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeProductUpdated,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceProductService,
		},
		Payload: ProductUpdatedPayload{
			ProductID: productID,
			Name:      name,
			Price:     price,
			Quantity:  quantity,
		},
	}
}

// ProductUpdatedProducer is responsible for producing product Updated events.
type ProductUpdatedProducer struct {
	Producer  *mq.KafkaAsyncProducer
	RetryChan chan *sarama.ProducerMessage
	topic     string
}

// NewProductUpdatedProducer creates a new instance of ProductUpdatedProducer.
func NewProductUpdatedProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducer {
	pr := &ProductUpdatedProducer{
		Producer:  producer,
		topic:     constant.ProductLifecycleTopic,
		RetryChan: make(chan *sarama.ProducerMessage, 100),
	}

	return pr
}

// Send implements the KafkaProducer interface.
func (p *ProductUpdatedProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *ProductUpdatedProducer) Topic() string {
	return p.topic
}
