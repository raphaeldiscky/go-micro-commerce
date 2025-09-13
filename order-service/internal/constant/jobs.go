package constant

import "time"

const (
	// JobRecoveryInterval is the interval at which the recovery job runs.
	JobRecoveryInterval = 5 * time.Minute
	// JobRecoveryMaxRetries is the maximum number of retries for the recovery job.
	JobRecoveryMaxRetries = 5
	// JobRecoveryMaxAge is the maximum age for which a saga can be recovered.
	JobRecoveryMaxAge = 24 * time.Hour
	// JobRecoveryTimeout is the timeout for the recovery job.
	JobRecoveryTimeout = 30 * time.Second
	// JobRecoveryMaxRowsFetch is the maximum number of rows to fetch at once.
	JobRecoveryMaxRowsFetch = 100
	// JobRecoveryRedisLockTTL is the TTL for the Redis distributed lock.
	JobRecoveryRedisLockTTL = 10 * time.Minute
	// JobRecoveryRedisLockBackoff is the backoff interval for Redis distributed lock.
	JobRecoveryRedisLockBackoff = 100 * time.Millisecond
	// JobRecoveryRedisLockMaxRetries is the maximum number of retries for Redis distributed lock.
	JobRecoveryRedisLockMaxRetries = 10
)
