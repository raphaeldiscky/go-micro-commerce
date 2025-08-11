package event

import (
	"context"
	"time"

	"github.com/IBM/sarama"
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
	metadata KafkaMetadata
	payload  ProductDeletedPayload
}

// GetPayload returns the data associated with the ProductDeletedEvent.
func (e *ProductDeletedEvent) GetPayload() interface{} {
	return e.payload
}

// GetMetadata returns the metadata associated with the ProductDeletedEvent.
func (e *ProductDeletedEvent) GetMetadata() KafkaMetadata {
	return e.metadata
}

// NewProductDeletedEvent creates a new ProductDeletedEvent.
func NewProductDeletedEvent(productID uuid.UUID) *ProductDeletedEvent {
	return &ProductDeletedEvent{
		metadata: KafkaMetadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeProductDeleted,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceProductService,
		},
		payload: ProductDeletedPayload{
			ProductID: productID,
		},
	}
}

// ProductDeletedProducer is responsible for producing product deleted events.
type ProductDeletedProducer struct {
	Producer  *mq.KafkaAsyncProducer
	RetryChan chan *sarama.ProducerMessage
	topic     string
}

// NewProductDeletedProducer creates a new instance of ProductDeletedProducer.
func NewProductDeletedProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducer {
	pr := &ProductDeletedProducer{
		Producer:  producer,
		topic:     constant.ProductLifecycleTopic,
		RetryChan: make(chan *sarama.ProducerMessage, 100),
	}

	return pr
}

// Send implements the KafkaProducer interface.
func (p *ProductDeletedProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *ProductDeletedProducer) Topic() string {
	return p.topic
}
