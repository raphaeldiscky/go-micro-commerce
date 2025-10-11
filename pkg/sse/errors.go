package sse

import "errors"

var (
	// ErrHubShutdown is returned when hub is shutting down.
	ErrHubShutdown = errors.New("hub is shutting down")

	// ErrConnectionClosed is returned when connection is closed.
	ErrConnectionClosed = errors.New("connection is closed")

	// ErrInvalidMessage is returned when message is invalid.
	ErrInvalidMessage = errors.New("invalid message")
)
