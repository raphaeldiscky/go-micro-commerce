package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// subscriber implements the Subscriber interface.
type subscriber struct {
	client   PubSubClient
	config   PubSubConfig
	logger   logger.Logger
	pubsub   *redis.PubSub
	mu       sync.RWMutex
	running  bool
	handlers map[string]MessageHandler // channel → handler mapping
}

// NewSubscriber creates a new Redis subscriber.
func NewSubscriber(client PubSubClient, config PubSubConfig, logger logger.Logger) Subscriber {
	return &subscriber{
		client:   client,
		config:   config,
		logger:   logger,
		handlers: make(map[string]MessageHandler),
	}
}

// Subscribe subscribes to one or more channels and calls the handler for each message.
// If already running, adds the new channels to the existing subscription.
func (s *subscriber) Subscribe(
	ctx context.Context,
	handler MessageHandler,
	channels ...string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store handler for each channel
	for _, channel := range channels {
		s.handlers[channel] = handler
	}

	if s.running {
		// Add channels to existing subscription
		if s.pubsub == nil {
			return errors.New("subscriber is running but pubsub is nil")
		}

		if err := s.pubsub.Subscribe(ctx, channels...); err != nil {
			// Remove handlers on error
			for _, channel := range channels {
				delete(s.handlers, channel)
			}

			return fmt.Errorf("failed to add channels to existing subscription: %w", err)
		}

		s.logger.Infof("Added channels to existing subscription: %v", channels)

		return nil
	}

	// Create new subscription
	s.pubsub = s.client.Subscribe(ctx, channels...)
	s.running = true

	go s.processMessages(ctx)

	s.logger.Infof("Subscribed to channels: %v", channels)

	return nil
}

// SubscribePattern subscribes to channels matching a pattern.
// If already running, adds the new pattern to the existing subscription.
func (s *subscriber) SubscribePattern(
	ctx context.Context,
	handler MessageHandler,
	pattern string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store handler for the pattern
	s.handlers[pattern] = handler

	if s.running {
		// Add pattern to existing subscription
		if s.pubsub == nil {
			return errors.New("subscriber is running but pubsub is nil")
		}

		if err := s.pubsub.PSubscribe(ctx, pattern); err != nil {
			// Remove handler on error
			delete(s.handlers, pattern)

			return fmt.Errorf("failed to add pattern to existing subscription: %w", err)
		}

		s.logger.Infof("Added pattern to existing subscription: %s", pattern)

		return nil
	}

	// Create new subscription
	s.pubsub = s.client.PSubscribe(ctx, pattern)
	s.running = true

	go s.processMessages(ctx)

	s.logger.Infof("Subscribed to pattern: %s", pattern)

	return nil
}

// processMessages processes incoming messages in a separate goroutine.
func (s *subscriber) processMessages(ctx context.Context) {
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	ch := s.pubsub.Channel(redis.WithChannelSize(s.config.ChannelBufferSize))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Context cancelled, stopping message processing")
			return
		case msg, ok := <-ch:
			if !ok {
				s.logger.Info("Channel closed, stopping message processing")
				return
			}

			if err := s.handleMessage(ctx, msg); err != nil {
				s.logger.Errorf("Failed to handle message: %v", err)
			}
		}
	}
}

// handleMessage processes a single Redis message.
func (s *subscriber) handleMessage(
	ctx context.Context,
	redisMsg *redis.Message,
) error {
	// Look up handler for this channel
	s.mu.RLock()
	handler, exists := s.handlers[redisMsg.Channel]
	s.mu.RUnlock()

	if !exists {
		// For pattern subscriptions, try to find matching pattern
		s.mu.RLock()

		for pattern, h := range s.handlers {
			if redisMsg.Pattern != "" && pattern == redisMsg.Pattern {
				handler = h
				exists = true

				break
			}
		}

		s.mu.RUnlock()

		if !exists {
			return fmt.Errorf("no handler found for channel: %s", redisMsg.Channel)
		}
	}

	message, err := FromJSON([]byte(redisMsg.Payload))
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	if handlerErr := handler(ctx, message); handlerErr != nil {
		return fmt.Errorf("handler failed for message %s: %w", message.GetMessageID(), handlerErr)
	}

	return nil
}

// Unsubscribe unsubscribes from specified channels.
func (s *subscriber) Unsubscribe(channels ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pubsub == nil {
		return errors.New("not subscribed to any channels")
	}

	err := s.pubsub.Unsubscribe(context.Background(), channels...)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from channels %v: %w", channels, err)
	}

	// Remove handlers for unsubscribed channels
	for _, channel := range channels {
		delete(s.handlers, channel)
	}

	s.logger.Infof("Unsubscribed from channels: %v", channels)

	return nil
}

// Close closes the subscriber and releases resources.
func (s *subscriber) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pubsub != nil {
		err := s.pubsub.Close()
		if err != nil {
			return fmt.Errorf("failed to close pubsub: %w", err)
		}

		s.pubsub = nil
	}

	// Clear all handlers
	s.handlers = make(map[string]MessageHandler)

	s.running = false
	s.logger.Info("Subscriber closed")

	return nil
}
