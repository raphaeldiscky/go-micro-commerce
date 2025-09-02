package constant

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
