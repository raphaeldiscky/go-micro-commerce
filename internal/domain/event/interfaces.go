package event

import "context"

// Publisher defines the interface for publishing domain event.
type Publisher interface {
	Publish(ctx context.Context, event DomainEvent) error
	PublishBatch(ctx context.Context, event []DomainEvent) error
}

// Subscriber defines the interface for subscribing to domain event.
type Subscriber interface {
	Subscribe(ctx context.Context, eventType string, handler Handler) error
	Unsubscribe(ctx context.Context, eventType string) error
}

// Handler defines the function signature for handling event.
type Handler func(ctx context.Context, event DomainEvent) error

// Bus combines publisher and subscriber interfaces.
type Bus interface {
	Publisher
	Subscriber
}
