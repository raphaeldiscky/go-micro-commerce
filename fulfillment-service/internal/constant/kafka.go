package constant

const (
	// PaymentLifecycleTopicNumPartitions is the number of partitions for the payment lifecycle topic.
	PaymentLifecycleTopicNumPartitions = 3
	// PaymentLifecycleTopicReplicationFactor is the replication factor for the payment lifecycle topic.
	PaymentLifecycleTopicReplicationFactor = 1
	// PaymentDLQTopicNumPartitions is the number of partitions for the payment DLQ topic.
	PaymentDLQTopicNumPartitions = 1
	// PaymentDLQTopicReplicationFactor is the replication factor for the payment DLQ topic.
	PaymentDLQTopicReplicationFactor = 1
)

// Kafka Topic Configuration Constants.
const (
	// FulfillmentLifecycleTopicNumPartitions defines the number of partitions for fulfillment lifecycle topic.
	FulfillmentLifecycleTopicNumPartitions = 3
	// FulfillmentLifecycleTopicReplicationFactor defines the replication factor for fulfillment lifecycle topic.
	FulfillmentLifecycleTopicReplicationFactor = 1
	// FulfillmentDLQTopicNumPartitions defines the number of partitions for fulfillment DLQ topic.
	FulfillmentDLQTopicNumPartitions = 3
	// FulfillmentDLQTopicReplicationFactor defines the replication factor for fulfillment DLQ topic.
	FulfillmentDLQTopicReplicationFactor = 1
)
