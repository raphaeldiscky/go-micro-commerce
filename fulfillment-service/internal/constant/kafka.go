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
