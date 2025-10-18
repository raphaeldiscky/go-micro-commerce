// Package eventbus provides a universal event bus abstraction for pub/sub messaging.
package eventbus

import "context"

// EventBus defines the interface for publishing and subscribing to events.
type EventBus interface {
	// Subscribe subscribes to a specific channel with a handler.
	// Multiple handlers can be registered for the same channel.
	Subscribe(channel string, handler EventHandler) error

	// Unsubscribe unsubscribes from a specific channel.
	Unsubscribe(channel string) error

	// Publish publishes an event to a channel.
	Publish(ctx context.Context, channel string, event Event) error

	// SSubscribe subscribes to a specific sharded channel with a handler (Redis 7.0+).
	// Sharded subscriptions use slot-based distribution for better scalability.
	SSubscribe(channel string, handler EventHandler) error

	// SUnsubscribe unsubscribes from a specific sharded channel.
	SUnsubscribe(channel string) error

	// SPublish publishes an event to a sharded channel (Redis 7.0+).
	// Sharded pub/sub uses slot-based distribution for better scalability.
	SPublish(ctx context.Context, channel string, event Event) error

	// Close closes the event bus and releases resources.
	Close() error
}

// EventHandler is a function that handles received events.
type EventHandler func(ctx context.Context, event Event) error

// Event defines the interface for events.
type Event interface {
	// GetSourceInstanceID returns the ID of the instance that published this event.
	GetSourceInstanceID() string

	// GetType returns the event type.
	GetType() string

	// Marshal serializes the event to bytes.
	Marshal() ([]byte, error)

	// UnmarshalPayload deserializes the event payload into the target.
	UnmarshalPayload(target interface{}) error
}
