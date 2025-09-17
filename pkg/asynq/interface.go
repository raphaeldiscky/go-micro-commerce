package asynq

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

// ClientInterface defines the interface for Asynq client operations.
type ClientInterface interface {
	// Enqueue enqueues a task to be processed immediately.
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
	// EnqueueIn enqueues a task to be processed after the given delay.
	EnqueueIn(d time.Duration, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
	// EnqueueAt enqueues a task to be processed at the given time.
	EnqueueAt(t time.Time, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
	// Close closes the client connection.
	Close() error
}

// ServerInterface defines the interface for Asynq server operations.
type ServerInterface interface {
	// Start starts the task server.
	Start(handler asynq.Handler) error
	// Stop stops the task server.
	Stop()
	// Shutdown gracefully shuts down the server.
	Shutdown()
}

// HandlerFunc defines the function signature for task handlers.
type HandlerFunc func(context.Context, *asynq.Task) error

// TaskPayload defines the interface for task payloads.
type TaskPayload interface {
	// TaskType returns the task type identifier.
	TaskType() string
	// Validate validates the payload.
	Validate() error
}
