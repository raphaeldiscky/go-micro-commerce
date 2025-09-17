package asynq

import "errors"

var (
	// ErrInvalidRedisAddr is returned when Redis address is invalid.
	ErrInvalidRedisAddr = errors.New("invalid redis address")
	// ErrInvalidConcurrency is returned when concurrency setting is invalid.
	ErrInvalidConcurrency = errors.New("invalid concurrency setting")
	// ErrInvalidQueues is returned when queue configuration is invalid.
	ErrInvalidQueues = errors.New("invalid queue configuration")
	// ErrInvalidMaxRetry is returned when max retry setting is invalid.
	ErrInvalidMaxRetry = errors.New("invalid max retry setting")
	// ErrClientNotInitialized is returned when client is not properly initialized.
	ErrClientNotInitialized = errors.New("asynq client not initialized")
	// ErrServerNotInitialized is returned when server is not properly initialized.
	ErrServerNotInitialized = errors.New("asynq server not initialized")
	// ErrInvalidTaskPayload is returned when task payload is invalid.
	ErrInvalidTaskPayload = errors.New("invalid task payload")
	// ErrTaskHandlerNotFound is returned when task handler is not found.
	ErrTaskHandlerNotFound = errors.New("task handler not found")
)
