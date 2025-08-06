package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
)

// NewProductCreatedEvent creates a new ProductCreatedEvent.
func NewProductCreatedEvent(
	productID uuid.UUID,
	name string,
	price decimal.Decimal,
) *ProductCreatedEvent {
	return &ProductCreatedEvent{
		BaseEvent: BaseEvent{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeProductCreated,
			AggregateID: productID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceProductService,
		},
		Data: ProductCreatedData{
			ProductID: productID,
			Name:      name,
			Price:     price,
		},
	}
}
