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
