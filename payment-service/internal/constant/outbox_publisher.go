package constant

import "time"

const (
	// OutboxBatchSize is the maximum number of events to process in a single batch.
	OutboxBatchSize = 100
	// OutboxPollInterval is the interval at which the outbox service polls for events to process.
	OutboxPollInterval = 5 * time.Second
	// OutboxMaxRetryAttempts is the maximum number of times to retry processing an event.
	OutboxMaxRetryAttempts = 5
	// OutboxRetryBackoff is the time to wait between retry attempts.
	OutboxRetryBackoff = 30 * time.Second
	// OutboxCleanupInterval is the interval at which the outbox service cleans up processed events.
	OutboxCleanupInterval = 1 * time.Hour
	// OutboxRetentionPeriod is the time-to-live for processed events.
	OutboxRetentionPeriod = 24 * time.Hour
)
