package kafka

const (
	// ProductLifecycleTopic is the topic for product lifecycle events.
	ProductLifecycleTopic = "product.lifecycle" // ProductCreated, ProductUpdated, ProductDeleted
)

const (
	// ProductCreatedEventType is the event type for product created events.
	ProductCreatedEventType = "ProductCreated"
	// ProductUpdatedEventType is the event type for product updated events.
	ProductUpdatedEventType = "ProductUpdated"
	// ProductDeletedEventType is the event type for product deleted events.
	ProductDeletedEventType = "ProductDeleted"
)

const (
	// ProductOrderEventsConsumerGroup is the consumer group for order lifecycle events.
	ProductOrderEventsConsumerGroup = "product-service.order-events"
)
