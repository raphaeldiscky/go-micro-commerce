package constant

const (
	// JobPaymentTimeoutBatchSize is the default batch size for payment timeout job.
	JobPaymentTimeoutBatchSize = 100
	// JobPaymentTimeoutRedisLockMaxRetries is the maximum number of retries for payment timeout job.
	JobPaymentTimeoutRedisLockMaxRetries = 3
)
