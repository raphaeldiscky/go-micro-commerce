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
	// CheckoutSessionLifecycleTopicNumPartitions is the number of partitions for the CheckoutSession lifecycle .
	CheckoutSessionLifecycleTopicNumPartitions = 3
	// CheckoutSessionLifecycleTopicReplicationFactor is the replication factor for the CheckoutSession lifecycle .
	CheckoutSessionLifecycleTopicReplicationFactor = 1
)
