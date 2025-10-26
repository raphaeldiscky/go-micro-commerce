package producer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// FulfillmentRequestEvent is the envelope for fulfillment request events.
type FulfillmentRequestEvent struct {
	Metadata kafkaevent.Metadata                  `json:"metadata"`
	Payload  kafkaevent.FulfillmentRequestPayload `json:"payload"`
}

// FulfillmentRequestProducer is responsible for producing Fulfillment Request events.
type FulfillmentRequestProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewFulfillmentRequestEvent creates a new FulfillmentRequestEvent.
func NewFulfillmentRequestEvent(
	order *entity.Order,
) *FulfillmentRequestEvent {
	// Convert order items to fulfillment items
	fulfillmentItems := make([]kafkaevent.FulfillmentItemPayload, len(order.Items))
	for i := range order.Items {
		fulfillmentItems[i] = kafkaevent.FulfillmentItemPayload{
			ProductID: order.Items[i].ProductID,
			Quantity:  order.Items[i].Quantity,
		}
	}

	return &FulfillmentRequestEvent{
		Metadata: kafkaevent.Metadata{
			EventID:     uuid.New(),
			EventType:   kafka.FulfillmentRequestedEventType,
			AggregateID: order.ID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.OrderServiceName,
		},
		Payload: kafkaevent.FulfillmentRequestPayload{
			OrderID:    order.ID,
			CustomerID: order.CustomerID,
			Currency:   order.Currency,
			Items:      fulfillmentItems,
			Courier: kafkaevent.Courier{
				CourierID: order.Courier.CourierID,
			},

			Destination: kafkaevent.Destination{
				Country:    order.Destination.Country,
				City:       order.Destination.City,
				State:      order.Destination.State,
				PostalCode: order.Destination.PostalCode,
			},
			Origin: kafkaevent.Origin{
				Country:    order.Origin.Country,
				City:       order.Origin.City,
				State:      order.Origin.State,
				PostalCode: order.Origin.PostalCode,
			},
			Package: kafkaevent.Package{
				WeightKG: order.Package.WeightKG,
				Length:   order.Package.Length,
				Height:   order.Package.Height,
				Width:    order.Package.Width,
				Unit:     order.Package.Unit,
			},
		},
	}
}

// GetPayload returns the data associated with the FulfillmentRequestEvent.
func (e *FulfillmentRequestEvent) GetPayload() any {
	return e.Payload
}

// GetMetadata returns the metadata associated with the FulfillmentRequestEvent.
func (e *FulfillmentRequestEvent) GetMetadata() kafkaevent.Metadata {
	return e.Metadata
}

// NewFulfillmentRequestProducer creates a new instance of FulfillmentRequestProducer.
func NewFulfillmentRequestProducer(producer *kafka.AsyncProducer) kafka.Producer {
	return &FulfillmentRequestProducer{
		Producer: producer,
		topic:    kafka.FulfillmentRequestTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *FulfillmentRequestProducer) Send(ctx context.Context, evt kafkaevent.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *FulfillmentRequestProducer) Topic() string {
	return p.topic
}
