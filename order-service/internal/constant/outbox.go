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

// OutboxStatus represents the status of an outbox event.
type OutboxStatus string

const (
	// OutboxStatusPending indicates that the event is pending and has not yet been processed.
	OutboxStatusPending OutboxStatus = "pending"
	// OutboxStatusProcessing indicates that the event is currently being processed.
	OutboxStatusProcessing OutboxStatus = "processing"
	// OutboxStatusProcessed indicates that the event has been processed successfully.
	OutboxStatusProcessed OutboxStatus = "processed"
	// OutboxStatusFailed indicates that the event has failed processing.
	OutboxStatusFailed OutboxStatus = "failed"
	// OutboxStatusRetry indicates that the event is scheduled for retry.
	OutboxStatusRetry OutboxStatus = "retry"
)
