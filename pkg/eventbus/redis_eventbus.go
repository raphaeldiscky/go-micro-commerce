package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	redispkg "github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
)

// redisEventBus implements EventBus using Redis pub/sub.
type redisEventBus struct {
	publisher       redispkg.Publisher
	subscriber      redispkg.Subscriber
	logger          logger.Logger
	instanceID      string
	subscriptions   map[string][]EventHandler // channel → handlers
	activeChannels  map[string]bool           // channels currently subscribed in Redis
	mutex           sync.RWMutex
	subscriberMutex sync.Mutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewRedisEventBus creates a new Redis-based event bus.
func NewRedisEventBus(
	publisher redispkg.Publisher,
	subscriber redispkg.Subscriber,
	instanceID string,
	logger logger.Logger,
) EventBus {
	ctx, cancel := context.WithCancel(context.Background())

	return &redisEventBus{
		publisher:      publisher,
		subscriber:     subscriber,
		logger:         logger,
		instanceID:     instanceID,
		subscriptions:  make(map[string][]EventHandler),
		activeChannels: make(map[string]bool),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Subscribe subscribes to a channel with an event handler.
func (b *redisEventBus) Subscribe(channel string, handler EventHandler) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Add handler to subscriptions
	b.subscriptions[channel] = append(b.subscriptions[channel], handler)

	// Subscribe to Redis channel if first subscription
	if len(b.subscriptions[channel]) == 1 {
		if err := b.subscribeToRedis(channel); err != nil {
			// Remove handler on error
			b.subscriptions[channel] = b.subscriptions[channel][:len(b.subscriptions[channel])-1]
			if len(b.subscriptions[channel]) == 0 {
				delete(b.subscriptions, channel)
			}

			return fmt.Errorf("failed to subscribe to Redis channel %s: %w", channel, err)
		}

		b.activeChannels[channel] = true
		b.logger.Info("Subscribed to channel", "channel", channel)
	}

	return nil
}

// Unsubscribe unsubscribes from a channel.
func (b *redisEventBus) Unsubscribe(channel string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Remove all handlers for this channel
	delete(b.subscriptions, channel)

	// Unsubscribe from Redis if active
	if b.activeChannels[channel] {
		if err := b.unsubscribeFromRedis(channel); err != nil {
			b.logger.Error("Failed to unsubscribe from Redis channel",
				"channel", channel,
				"error", err)
			// Continue even on error to clean up local state
		}

		delete(b.activeChannels, channel)
		b.logger.Info("Unsubscribed from channel", "channel", channel)
	}

	return nil
}

// Publish publishes an event to a channel.
func (b *redisEventBus) Publish(ctx context.Context, channel string, event Event) error {
	// Serialize event
	data, err := event.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Redis message
	metadata := redispkg.NewMessageMetadata("chat-service")

	redisMsg, err := redispkg.NewMessage(metadata, json.RawMessage(data))
	if err != nil {
		return fmt.Errorf("failed to create Redis message: %w", err)
	}

	// Publish to Redis
	if err = b.publisher.Publish(ctx, channel, redisMsg); err != nil {
		return fmt.Errorf("failed to publish to channel %s: %w", channel, err)
	}

	b.logger.Debug("Published event to channel",
		"channel", channel,
		"event_type", event.GetType(),
		"instance_id", event.GetSourceInstanceID())

	return nil
}

// SSubscribe subscribes to a sharded channel with an event handler (Redis 7.0+).
// Sharded subscriptions use slot-based distribution for better scalability in Redis Cluster.
func (b *redisEventBus) SSubscribe(channel string, handler EventHandler) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Add handler to subscriptions
	b.subscriptions[channel] = append(b.subscriptions[channel], handler)

	// Subscribe to Redis sharded channel if first subscription
	if len(b.subscriptions[channel]) == 1 {
		if err := b.ssubscribeToRedis(channel); err != nil {
			// Remove handler on error
			b.subscriptions[channel] = b.subscriptions[channel][:len(b.subscriptions[channel])-1]
			if len(b.subscriptions[channel]) == 0 {
				delete(b.subscriptions, channel)
			}

			return fmt.Errorf("failed to subscribe to Redis sharded channel %s: %w", channel, err)
		}

		b.activeChannels[channel] = true
		b.logger.Info("Subscribed to sharded channel", "channel", channel)
	}

	return nil
}

// SUnsubscribe unsubscribes from a sharded channel.
func (b *redisEventBus) SUnsubscribe(channel string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Remove all handlers for this channel
	delete(b.subscriptions, channel)

	// Unsubscribe from Redis if active
	if b.activeChannels[channel] {
		if err := b.sunsubscribeFromRedis(channel); err != nil {
			b.logger.Error("Failed to unsubscribe from Redis sharded channel",
				"channel", channel,
				"error", err)
			// Continue even on error to clean up local state
		}

		delete(b.activeChannels, channel)
		b.logger.Info("Unsubscribed from sharded channel", "channel", channel)
	}

	return nil
}

// SPublish publishes an event to a sharded channel (Redis 7.0+).
// Sharded pub/sub uses slot-based distribution for better scalability in Redis Cluster.
func (b *redisEventBus) SPublish(ctx context.Context, channel string, event Event) error {
	// Serialize event
	data, err := event.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Redis message
	metadata := redispkg.NewMessageMetadata("eventbus")

	redisMsg, err := redispkg.NewMessage(metadata, json.RawMessage(data))
	if err != nil {
		return fmt.Errorf("failed to create Redis message: %w", err)
	}

	// Publish to Redis sharded channel
	if err = b.publisher.SPublish(ctx, channel, redisMsg); err != nil {
		return fmt.Errorf("failed to publish to sharded channel %s: %w", channel, err)
	}

	b.logger.Debug("Published event to sharded channel",
		"channel", channel,
		"event_type", event.GetType(),
		"instance_id", event.GetSourceInstanceID())

	return nil
}

// ssubscribeToRedis subscribes to a Redis sharded channel (must be called with lock held).
func (b *redisEventBus) ssubscribeToRedis(channel string) error {
	b.subscriberMutex.Lock()
	defer b.subscriberMutex.Unlock()

	// Create a handler that routes to registered handlers
	redisHandler := func(ctx context.Context, redisMsg *redispkg.Message) error {
		return b.handleRedisMessage(ctx, channel, redisMsg)
	}

	// Subscribe to Redis sharded channel
	if err := b.subscriber.SSubscribe(b.ctx, redisHandler, channel); err != nil {
		return err
	}

	return nil
}

// sunsubscribeFromRedis unsubscribes from a Redis sharded channel (must be called with lock held).
func (b *redisEventBus) sunsubscribeFromRedis(channel string) error {
	b.subscriberMutex.Lock()
	defer b.subscriberMutex.Unlock()

	return b.subscriber.SUnsubscribe(channel)
}

// Close closes the event bus and releases resources.
func (b *redisEventBus) Close() error {
	b.cancel()

	var errs []error

	if err := b.subscriber.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close subscriber: %w", err))
	}

	if err := b.publisher.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close publisher: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("event bus close errors: %v", errs)
	}

	b.logger.Info("Event bus closed successfully")

	return nil
}

// subscribeToRedis subscribes to a Redis channel (must be called with lock held).
func (b *redisEventBus) subscribeToRedis(channel string) error {
	b.subscriberMutex.Lock()
	defer b.subscriberMutex.Unlock()

	// Create a handler that routes to registered handlers
	redisHandler := func(ctx context.Context, redisMsg *redispkg.Message) error {
		return b.handleRedisMessage(ctx, channel, redisMsg)
	}

	// Subscribe to Redis channel
	if err := b.subscriber.Subscribe(b.ctx, redisHandler, channel); err != nil {
		return err
	}

	return nil
}

// unsubscribeFromRedis unsubscribes from a Redis channel (must be called with lock held).
func (b *redisEventBus) unsubscribeFromRedis(channel string) error {
	b.subscriberMutex.Lock()
	defer b.subscriberMutex.Unlock()

	return b.subscriber.Unsubscribe(channel)
}

// handleRedisMessage handles incoming Redis messages.
func (b *redisEventBus) handleRedisMessage(
	ctx context.Context,
	channel string,
	redisMsg *redispkg.Message,
) error {
	// Unmarshal event
	var rawEvent json.RawMessage
	if err := redisMsg.UnmarshalPayload(&rawEvent); err != nil {
		return fmt.Errorf("failed to unmarshal Redis message: %w", err)
	}

	event, err := Unmarshal([]byte(rawEvent))
	if err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Note: Not filtering by instance ID - application handlers decide whether to process
	b.logger.Debug("Received event",
		"channel", channel,
		"event_type", event.EventType,
		"source_instance", event.SourceInstanceID,
		"our_instance", b.instanceID)

	// Get handlers for this channel
	b.mutex.RLock()
	handlers, exists := b.subscriptions[channel]
	handlersCopy := make([]EventHandler, len(handlers))
	copy(handlersCopy, handlers)
	b.mutex.RUnlock()

	if !exists || len(handlersCopy) == 0 {
		b.logger.Warn("No handlers registered for channel", "channel", channel)
		return nil
	}

	// Call all handlers
	var errs []error

	for _, handler := range handlersCopy {
		if err = handler(ctx, event); err != nil {
			b.logger.Error("Event handler failed",
				"channel", channel,
				"event_type", event.EventType,
				"error", err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"one or more handlers failed for channel %s: %d errors",
			channel,
			len(errs),
		)
	}

	return nil
}
