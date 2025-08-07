package event

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
)

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
