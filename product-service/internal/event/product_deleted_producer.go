package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
)

// ProductDeletedPayload represents when a product is deleted.
type ProductDeletedPayload struct {
	ProductID uuid.UUID `json:"product_id"`
}

// ProductDeletedEvent is the envelope for product deletion events.
type ProductDeletedEvent struct {
	Metadata KafkaMetadata
	Payload  ProductDeletedPayload
}

// GetPayload returns the data associated with the ProductDeletedEvent.
func (e *ProductDeletedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the ProductDeletedEvent.
func (e *ProductDeletedEvent) GetMetadata() KafkaMetadata {
	return e.Metadata
}

// NewProductDeletedEvent creates a new ProductDeletedEvent.
func NewProductDeletedEvent(productID uuid.UUID) *ProductDeletedEvent {
	return &ProductDeletedEvent{
		Metadata: KafkaMetadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeProductDeleted,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceProductService,
		},
		Payload: ProductDeletedPayload{
			ProductID: productID,
		},
	}
}

// ProductDeletedProducer is responsible for producing product deleted events.
type ProductDeletedProducer struct {
	Producer *mq.KafkaAsyncProducer
	topic    string
}

// NewProductDeletedProducer creates a new instance of ProductDeletedProducer.
func NewProductDeletedProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducerInterface {
	return &ProductDeletedProducer{
		Producer: producer,
		topic:    constant.ProductLifecycleTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *ProductDeletedProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *ProductDeletedProducer) Topic() string {
	return p.topic
}
