package constant

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
