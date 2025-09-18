package constant

import "time"

const (
	// DefaultAsynqConcurrency is the default number of concurrent workers.
	DefaultAsynqConcurrency = 100
	// DefaultAsynqMaxRetry is the default maximum number of retries.
	DefaultAsynqMaxRetry = 0

	// QueuePriorityCritical is the priority distribution for critical tasks.
	QueuePriorityCritical = 6 // 60%
	// QueuePriorityDefault is the priority distribution for default tasks.
	QueuePriorityDefault = 3
	// QueuePriorityLow is the priority distribution for low priority tasks.
	QueuePriorityLow = 1

	// DefaultRetryDelay is the default delay between retries.
	DefaultRetryDelay = 5 * time.Second
	// DefaultRetryMaxDelay is the maximum delay between retries.
	DefaultRetryMaxDelay = 5 * time.Minute

	// DefaultHealthCheckInterval is the default health check interval.
	DefaultHealthCheckInterval = 15 * time.Second
	// DefaultDelayedTaskCheckInterval is the default delayed task check interval.
	DefaultDelayedTaskCheckInterval = 5 * time.Second

	// TaskRetentionHours is the retention period for completed tasks.
	TaskRetentionHours = 24 * time.Hour
)
