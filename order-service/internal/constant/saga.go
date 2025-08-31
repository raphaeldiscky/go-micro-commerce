package constant

// SagaStatus represents the status of a saga execution.
type SagaStatus string

const (
	// SagaStatusPending indicates that the saga is pending.
	SagaStatusPending SagaStatus = "pending"
	// SagaStatusExecuting indicates that the saga is currently executing.
	SagaStatusExecuting SagaStatus = "executing"
	// SagaStatusCompensating indicates that the saga is compensating due to a failure.
	SagaStatusCompensating SagaStatus = "compensating"
	// SagaStatusCompleted indicates that the saga has completed successfully.
	SagaStatusCompleted SagaStatus = "completed"
	// SagaStatusFailed indicates that the saga has failed.
	SagaStatusFailed SagaStatus = "failed"
	// SagaStatusCompensated indicates that the saga has been compensated after a failure.
	SagaStatusCompensated SagaStatus = "compensated"
)
