package constant

import "time"

const (
	// TemporalRetryInterval is the retry interval for Temporal tasks.
	TemporalRetryInterval = 500 * time.Millisecond
	// TemporalBackoffCoefficient is the backoff coefficient for Temporal tasks.
	TemporalBackoffCoefficient = 2.0
	// TemporalMaxAttempts is the maximum number of attempts for Temporal tasks.
	TemporalMaxAttempts = 3
	// TemporalMaxInterval is the maximum interval for Temporal tasks.
	TemporalMaxInterval = 1 * time.Minute
	// TemporalWorkflowTimeout is the start-to-close timeout for Temporal tasks.
	TemporalWorkflowTimeout = 90 * time.Minute
	// TemporalCompensationWorkflowTimeout is the start-to-close timeout for compensation Temporal tasks.
	TemporalCompensationWorkflowTimeout = 15 * time.Minute
)
