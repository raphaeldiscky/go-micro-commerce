package event

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-template/services/product-service/internal/constant"
)

// NewProductDeletedEvent creates a new ProductDeletedEvent.
func NewProductDeletedEvent(productID uuid.UUID) *ProductDeletedEvent {
	return &ProductDeletedEvent{
		BaseEvent: BaseEvent{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeProductDeleted,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceProductService,
		},
		Data: ProductDeletedData{
			ProductID: productID,
		},
	}
}

// GetData returns the data associated with the ProductDeletedEvent.
func (e *ProductDeletedEvent) GetData() interface{} {
	return e.Data
}
