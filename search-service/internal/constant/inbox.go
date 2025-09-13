// Package constant contains constants related to inbox events.
package constant

import "time"

const (
	// InboxPollInterval is the interval at which the inbox processor will poll for events.
	InboxPollInterval = 5 * time.Second
	// InboxCleanupInterval is the interval at which the inbox processor will clean up processed events.
	InboxCleanupInterval = 24 * time.Hour
	// InboxRetentionPeriod is the time-to-live for processed events.
	InboxRetentionPeriod = 168 * time.Hour
	// InboxBatchSize is the maximum number of events to process in a single batch.
	InboxBatchSize = 100
	// InboxMaxRetryAttempts is the maximum number of retry attempts for a failed event.
	InboxMaxRetryAttempts = 3
	// InboxRetryBackoff is the backoff duration for retry attempts.
	InboxRetryBackoff = 5 * time.Second
)

// InboxStatus represents the status of an inbox event.
type InboxStatus string

const (
	// InboxStatusPending indicates that the event is pending.
	InboxStatusPending InboxStatus = "pending"
	// InboxStatusProcessing indicates that the event is being processed.
	InboxStatusProcessing InboxStatus = "processing"
	// InboxStatusProcessed indicates that the event has been processed.
	InboxStatusProcessed InboxStatus = "processed"
	// InboxStatusFailed indicates that the event has failed.
	InboxStatusFailed InboxStatus = "failed"
	// InboxStatusRetry indicates that the event is scheduled for retry.
	InboxStatusRetry InboxStatus = "retry"
	// InboxStatusDuplicate indicates that the event is a duplicate.
	InboxStatusDuplicate InboxStatus = "duplicate"
)
