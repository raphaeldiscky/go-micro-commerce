package constant

import "time"

const (
	// KafkaRetryMax is the maximum number of retries for Kafka messages.
	KafkaRetryMax = 3
	// KafkaRetryInterval is the interval between retries for Kafka messages.
	KafkaRetryInterval = 2 * time.Second
	// KafkaFlushFrequency is the frequency at which messages are flushed to Kafka.
	KafkaFlushFrequency = 1000
)

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
