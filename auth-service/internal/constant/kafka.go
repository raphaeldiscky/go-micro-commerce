// Package constant defines constants used in the auth service.
package constant

import "time"

const (
	// KafkaRetryMax is the maximum number of retries for Kafka operations.
	KafkaRetryMax = 3
	// KafkaRetryInterval is the interval between retries for Kafka operations.
	KafkaRetryInterval = 500 * time.Millisecond
	// KafkaFlushFrequency is the frequency at which messages are flushed to Kafka.
	KafkaFlushFrequency = 100
)

const (
	// UserVerificationTopicNumPartitions is the number of partitions for the user verification topic.
	UserVerificationTopicNumPartitions = 3
	// UserVerificationTopicReplicationFactor is the replication factor for the user verification topic.
	UserVerificationTopicReplicationFactor = 1
)
