package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-template/services/product-service/internal/constant"
)

// NewProductUpdatedEvent creates a new ProductUpdatedEvent.
func NewProductUpdatedEvent(
	productID uuid.UUID,
	name string,
	price decimal.Decimal,
) *ProductUpdatedEvent {
	return &ProductUpdatedEvent{
		BaseEvent: BaseEvent{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeProductUpdated,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceProductService,
		},
		Data: ProductUpdatedData{
			ProductID: productID,
			Name:      name,
			Price:     price,
		},
	}
}

// GetData returns the data associated with the ProductUpdatedEvent.
func (e *ProductUpdatedEvent) GetData() interface{} {
	return e.Data
}
