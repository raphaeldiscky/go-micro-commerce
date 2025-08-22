// Package constant defines constants used throughout the product service.
package constant

// Product Service Source.
const (
	// KafkaSourceProductService is the source identifier for events produced by the product service.
	KafkaSourceProductService = "product-service"
)

// Product Service Event Types.
const (
	// KafkaEventTypeProductCreated is the event type for product created events.
	KafkaEventTypeProductCreated = "ProductCreated"
	// KafkaEventTypeProductUpdated is the event type for product updated events.
	KafkaEventTypeProductUpdated = "ProductUpdated"
	// KafkaEventTypeProductDeleted is the event type for product deleted events.
	KafkaEventTypeProductDeleted = "ProductDeleted"
	// KafkaEventTypeProductDeleted is the event type for product deleted events.
)

// Topics that Product Service produces to.
const (
	// TopicProductLifecycle is the topic for product lifecycle events.
	TopicProductLifecycle = "product.lifecycle" // ProductCreated, ProductUpdated, ProductDeleted
	// TopicProductLifecycleNumPartitions is the number of partitions for the product lifecycle topic.
	TopicProductLifecycleNumPartitions = 3
	// TopicProductLifecycleReplicationFactor is the replication factor for the product lifecycle topic.
	TopicProductLifecycleReplicationFactor = 1
)
