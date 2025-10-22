package constant

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
