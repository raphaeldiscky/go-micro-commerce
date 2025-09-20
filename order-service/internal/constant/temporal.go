// Package constant provides Temporal workflow configuration constants.
package constant

import "time"

// Temporal Workflow Configuration
// These constants define timeouts, retry policies, and other configuration
// parameters for Temporal workflows and activities in the order processing saga.
const (
	// TemporalRetryInterval is the base retry interval for Temporal activities.
	// Used in config defaults and retry policies for failed activity executions.
	TemporalRetryInterval = 1 * time.Second

	// TemporalBackoffCoefficient controls exponential backoff for activity retries.
	// A value of 2.0 means each retry delay is doubled from the previous attempt.
	TemporalBackoffCoefficient = 2.0

	// TemporalMaxAttempts is the default maximum retry attempts for Temporal activities.
	// Set to 1 to disable retries by default (can be overridden per activity).
	TemporalMaxAttempts = 1

	// TemporalMaxInterval is the maximum delay between retry attempts.
	// Prevents exponential backoff from creating excessively long delays.
	TemporalMaxInterval = 1 * time.Minute

	// TemporalWorkflowTimeout is the maximum execution time for the entire order saga workflow.
	// This timeout covers all steps from product reservation to order confirmation.
	TemporalWorkflowTimeout = 25 * time.Hour // enough for all steps

	// TemporalCompensationWorkflowTimeout is the maximum time allowed for compensation activities.
	// Compensation includes refunds, product releases, and shipping cancellations.
	// Shorter than main workflow timeout since compensation should be faster.
	TemporalCompensationWorkflowTimeout = 15 * time.Minute
)

// Activity Timeout Aliases
// These constants provide convenient aliases to saga step timeouts for Temporal activities.
// They ensure consistency between saga and Temporal implementations.
const (
	// WaitForPaymentConfirmationActivityTimeout is the timeout for payment confirmation activity.
	// Matches the saga step timeout to maintain consistent behavior across implementations.
	WaitForPaymentConfirmationActivityTimeout = WaitForPaymentConfirmationStepTimeout
)
