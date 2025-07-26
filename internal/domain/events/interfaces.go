package events

import "context"

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event DomainEvent) error
	PublishBatch(ctx context.Context, events []DomainEvent) error
}

// EventSubscriber defines the interface for subscribing to domain events
type EventSubscriber interface {
	Subscribe(ctx context.Context, eventType string, handler EventHandler) error
	Unsubscribe(ctx context.Context, eventType string) error
}

// EventHandler defines the function signature for handling events
type EventHandler func(ctx context.Context, event DomainEvent) error

// EventBus combines publisher and subscriber interfaces
type EventBus interface {
	EventPublisher
	EventSubscriber
}
