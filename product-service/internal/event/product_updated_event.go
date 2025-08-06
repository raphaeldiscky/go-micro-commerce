package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
)

// NewProductUpdatedEvent creates a new ProductUpdatedEvent.
func NewProductUpdatedEvent(
	productID uuid.UUID,
	name string,
	price decimal.Decimal,
	quantity int,
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
			Quantity:  quantity,
		},
	}
}
