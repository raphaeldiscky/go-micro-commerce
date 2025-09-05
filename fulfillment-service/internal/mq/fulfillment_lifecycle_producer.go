package mq

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/entity"
)

// FulfillmentLifecycleEvent is the envelope for all Fulfillment events.
type FulfillmentLifecycleEvent struct {
	Metadata event.Metadata                    `json:"metadata"`
	Payload  event.FulfillmentLifecyclePayload `json:"payload"`
}

// GetPayload returns the data associated with the FulfillmentLifecycleEvent.
func (e *FulfillmentLifecycleEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the FulfillmentLifecycleEvent.
func (e *FulfillmentLifecycleEvent) GetMetadata() event.Metadata {
	return e.Metadata
}

// FulfillmentLifecycleProducer is responsible for producing Fulfillment Lifecycle events.
type FulfillmentLifecycleProducer struct {
	Producer *kafka.AsyncProducer
	topic    string
}

// NewFulfillmentLifecycleEvent creates a new FulfillmentLifecycleEvent.
func NewFulfillmentLifecycleEvent(
	fulfillment *entity.Fulfillment,
) *FulfillmentLifecycleEvent {
	return &FulfillmentLifecycleEvent{
		Metadata: event.Metadata{
			EventID:     uuid.New(),
			EventType:   getEventTypeFromStatus(fulfillment.Status),
			AggregateID: fulfillment.ID,
			OccurredAt:  time.Now().UTC(),
			Source:      pkgconstant.FulfillmentServiceName,
		},
		Payload: event.FulfillmentLifecyclePayload{
			FulfillmentID:       fulfillment.ID,
			OrderID:             fulfillment.OrderID,
			Status:              string(fulfillment.Status),
			ShippingCost:        fulfillment.ShippingCost,
			TrackingNumber:      fulfillment.TrackingNumber,
			EstimatedDeliveryAt: fulfillment.EstimatedDeliveryAt,
		},
	}
}

// NewFulfillmentLifecycleProducer creates a new instance of FulfillmentLifecycleProducer.
func NewFulfillmentLifecycleProducer(producer *kafka.AsyncProducer) kafka.ProducerInterface {
	return &FulfillmentLifecycleProducer{
		Producer: producer,
		topic:    kafka.FulfillmentLifecycleTopic,
	}
}

// Send implements the KafkaProducer interface.
func (p *FulfillmentLifecycleProducer) Send(ctx context.Context, evt event.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, evt)
}

// Topic returns the topic name.
func (p *FulfillmentLifecycleProducer) Topic() string {
	return p.topic
}

// getEventTypeFromStatus returns the appropriate event type based on fulfillment status.
func getEventTypeFromStatus(status constant.FulfillmentStatus) string {
	switch status {
	case constant.FulfillmentStatusPending:
		return kafka.FulfillmentCreatedEventType
	case constant.FulfillmentStatusProcessing:
		return kafka.FulfillmentProcessingEventType
	case constant.FulfillmentStatusShipped:
		return kafka.FulfillmentShippedEventType
	case constant.FulfillmentStatusInTransit:
		return kafka.FulfillmentInTransitEventType
	case constant.FulfillmentStatusDelivered:
		return kafka.FulfillmentDeliveredEventType
	case constant.FulfillmentStatusCanceled:
		return kafka.FulfillmentCanceledEventType
	case constant.FulfillmentStatusReturned:
		return kafka.FulfillmentReturnedEventType
	default:
		return kafka.FulfillmentUpdatedEventType
	}
}
