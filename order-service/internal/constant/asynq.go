package constant

import "time"

const (
	// DefaultAsynqConcurrency is the default number of concurrent workers.
	DefaultAsynqConcurrency = 10
	// DefaultAsynqMaxRetry is the default maximum number of retries.
	DefaultAsynqMaxRetry = 3

	// QueuePriorityCritical is the priority for critical tasks.
	QueuePriorityCritical = 6
	// QueuePriorityDefault is the priority for default tasks.
	QueuePriorityDefault = 3
	// QueuePriorityLow is the priority for low priority tasks.
	QueuePriorityLow = 1

	// DefaultRetryDelay is the default delay between retries.
	DefaultRetryDelay = 5 * time.Second
	// DefaultRetryMaxDelay is the maximum delay between retries.
	DefaultRetryMaxDelay = 5 * time.Minute

	// DefaultHealthCheckInterval is the default health check interval.
	DefaultHealthCheckInterval = 15 * time.Second
	// DefaultDelayedTaskCheckInterval is the default delayed task check interval.
	DefaultDelayedTaskCheckInterval = 5 * time.Second

	// FinalTaskDelayMinutes is the delay for final task scheduling.
	FinalTaskDelayMinutes = 10
	// CancelTaskDelayMinutes is the delay for cancel task scheduling.
	CancelTaskDelayMinutes = 20

	// FirstPaymentReminderMinutes is the delay for first task scheduling.
	FirstPaymentReminderMinutes = 0 * time.Minute
	// SecondPaymentReminderMinutes is the delay for second task scheduling.
	SecondPaymentReminderMinutes = 10 * time.Minute
	// CancelOrderDelayMinutes is the delay for cancel task scheduling.
	CancelOrderDelayMinutes = 15 * time.Minute
)

// GetDefaultAsynqQueues returns the default queue configuration for asynq.
func GetDefaultAsynqQueues() map[string]int {
	return map[string]int{
		"critical": QueuePriorityCritical, // High priority for urgent tasks
		"default":  QueuePriorityDefault,  // Normal priority
		"low":      QueuePriorityLow,      // Low priority for background tasks
	}
}
