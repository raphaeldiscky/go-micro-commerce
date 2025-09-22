package websocket

import "errors"

// Common WebSocket errors.
var (
	// ErrConnectionClosed indicates the connection is closed.
	ErrConnectionClosed = errors.New("websocket connection is closed")

	// ErrSendBufferFull indicates the send buffer is full.
	ErrSendBufferFull = errors.New("websocket send buffer is full")

	// ErrHubShutdown indicates the hub is shutting down.
	ErrHubShutdown = errors.New("websocket hub is shutting down")

	// ErrInvalidMessage indicates an invalid message format.
	ErrInvalidMessage = errors.New("invalid websocket message format")

	// ErrUnauthorized indicates unauthorized access.
	ErrUnauthorized = errors.New("unauthorized websocket access")

	// ErrChannelNotFound indicates a channel was not found.
	ErrChannelNotFound = errors.New("websocket channel not found")
)
