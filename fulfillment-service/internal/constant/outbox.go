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

// DLQReason represents the reason an event was sent to the dead-letter queue (DLQ).
type DLQReason string

const (
	// DLQReasonMaxRetriesExceeded indicates that the event has exceeded the maximum number of retry attempts.
	DLQReasonMaxRetriesExceeded DLQReason = "max_retries_exceeded"
	// DLQReasonDeserializationError indicates that the event failed to be deserialized.
	DLQReasonDeserializationError DLQReason = "deserialization_error"
	// DLQReasonValidationError indicates that the event failed validation.
	DLQReasonValidationError DLQReason = "validation_error"
	// DLQReasonProcessingTimeout indicates that the event processing timed out.
	DLQReasonProcessingTimeout DLQReason = "processing_timeout"
	// DLQReasonUnknownError indicates that an unknown error occurred.
	DLQReasonUnknownError DLQReason = "unknown_error"
)
