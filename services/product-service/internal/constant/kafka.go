package constant

const (
	// ProductCreatedTopic is the topic name for product created events.
	ProductCreatedTopic = "product.created"
	// ProductCreatedTopicNumPartitions is the number of partitions for the product created topic.
	ProductCreatedTopicNumPartitions = 3
	// ProductCreatedTopicReplicationFactor is the replication factor for the product created topic.
	ProductCreatedTopicReplicationFactor = 1

	// ProductUpdatedTopic is the topic name for product updated events.
	ProductUpdatedTopic = "product.updated"
	// ProductUpdatedTopicNumPartitions is the number of partitions for the product updated topic.
	ProductUpdatedTopicNumPartitions = 3
	// ProductUpdatedTopicReplicationFactor is the replication factor for the product updated topic.
	ProductUpdatedTopicReplicationFactor = 1

	// ProductDeletedTopic is the topic name for product deleted events.
	ProductDeletedTopic = "product.deleted"
	// ProductDeletedTopicNumPartitions is the number of partitions for the product deleted topic.
	ProductDeletedTopicNumPartitions = 3
	// ProductDeletedTopicReplicationFactor is the replication factor for the product deleted topic.
	ProductDeletedTopicReplicationFactor = 1
)

const (
	// KafkaProducerRetryDelay is the delay in seconds before retrying a failed Kafka message send.
	KafkaProducerRetryDelay = 3
	// KafkaProducerRetryLimit is the maximum number of retries for sending a message to Kafka.
	KafkaProducerRetryLimit = 3
	// KafkaConsumerRetryDelay is the delay in seconds before retrying a failed Kafka message processing.
	KafkaConsumerRetryDelay = 2
	// KafkaConsumerRetryLimit is the maximum number of retries for processing a message from Kafka.
	KafkaConsumerRetryLimit = 3
)

const (
	// KafkaEventTypeProductCreated is the event type for product created events.
	KafkaEventTypeProductCreated = "ProductCreated"
	// KafkaEventTypeProductUpdated is the event type for product updated events.
	KafkaEventTypeProductUpdated = "ProductUpdated"
	// KafkaEventTypeProductDeleted is the event type for product deleted events.
	KafkaEventTypeProductDeleted = "ProductDeleted"
	// KafkaEventTypeProductDeleted is the event type for product deleted events.
)

const (
	// KafkaSourceProductService is the source identifier for events produced by the product service.
	KafkaSourceProductService = "product-service"
)
