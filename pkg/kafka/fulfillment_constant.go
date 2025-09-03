package kafka

const (
	// FulfillmentRequestTopic is the topic for fulfillment events.
	FulfillmentRequestTopic = "fulfillment.request"
	// FulfillmentLifecycleTopic is the topic for fulfillment lifecycle events.
	FulfillmentLifecycleTopic = "fulfillment.lifecycle" // FulfillmentCreated, FulfillmentUpdated, FulfillmentCancelled, FulfillmentShipped, FulfillmentDelivered
)

const (
	// FulfillmentRequestedEventType is when fulfillment is requested for an order (pending).
	FulfillmentRequestedEventType = "FulfillmentRequested"
	// FulfillmentCreatedEventType is the event type for fulfillment created events.
	FulfillmentCreatedEventType = "FulfillmentCreated"
	// FulfillmentProcessingEventType is the event type for fulfillment processing events.
	FulfillmentProcessingEventType = "FulfillmentProcessing"
	// FulfillmentShippedEventType is the event type for fulfillment shipped events.
	FulfillmentShippedEventType = "FulfillmentShipped"
	// FulfillmentInTransitEventType is the event type for fulfillment in transit events.
	FulfillmentInTransitEventType = "FulfillmentInTransit"
	// FulfillmentDeliveredEventType is the event type for fulfillment delivered events.
	FulfillmentDeliveredEventType = "FulfillmentDelivered"
	// FulfillmentCanceledEventType is the event type for fulfillment canceled events.
	FulfillmentCanceledEventType = "FulfillmentCanceled"
	// FulfillmentReturnedEventType is the event type for fulfillment returned events.
	FulfillmentReturnedEventType = "FulfillmentReturned"
	// FulfillmentUpdatedEventType is the event type for fulfillment updated events.
	FulfillmentUpdatedEventType = "FulfillmentUpdated"
)

const (
	// FulfillmentOrderEventsConsumerGroup is the consumer group for order events.
	FulfillmentOrderEventsConsumerGroup = "fulfillment-service.order-events" // For order lifecycle
	// FulfillmentEventsConsumerGroup is the consumer group for fulfillment request events.
	FulfillmentEventsConsumerGroup = "fulfillment-service.fulfillment-events" // For fulfillment requests
	// OrderFulfillmentEventsConsumerGroup is the consumer group for order service consuming fulfillment events.
	OrderFulfillmentEventsConsumerGroup = "order-service.fulfillment-events" // For order service consuming fulfillment lifecycle
)
