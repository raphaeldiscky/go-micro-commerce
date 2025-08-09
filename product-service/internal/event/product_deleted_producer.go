package event

import (
	"time"

	"github.com/google/uuid"

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
