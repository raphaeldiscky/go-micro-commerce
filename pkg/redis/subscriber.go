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
	client         PubSubClient
	config         PubSubConfig
	logger         logger.Logger
	pubsub         *redis.PubSub            // For regular Subscribe/PSubscribe
	shardedPubsubs map[string]*redis.PubSub // For SSubscribe: channel → dedicated pubsub
	mu             sync.RWMutex
	running        bool
	shardedRunning map[string]bool           // track running state per sharded channel
	handlers       map[string]MessageHandler // channel → handler mapping
}

// NewSubscriber creates a new Redis subscriber.
func NewSubscriber(client PubSubClient, config PubSubConfig, logger logger.Logger) Subscriber {
	return &subscriber{
		client:         client,
		config:         config,
		logger:         logger,
		handlers:       make(map[string]MessageHandler),
		shardedPubsubs: make(map[string]*redis.PubSub),
		shardedRunning: make(map[string]bool),
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

	s.logger.Info("Redis subscriber message processing started",
		"buffer_size", s.config.ChannelBufferSize)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Context cancelled, stopping message processing")
			return
		case msg, ok := <-ch:
			if !ok {
				s.logger.Warn(
					"Redis pubsub channel closed unexpectedly, stopping message processing",
				)

				return
			}

			s.logger.Debug("Received message from Redis",
				"channel", msg.Channel,
				"pattern", msg.Pattern,
				"payload_size", len(msg.Payload))

			if err := s.handleMessage(ctx, msg); err != nil {
				s.logger.Errorf("Failed to handle message from Redis channel %s: %v",
					msg.Channel, err)
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
			s.logger.Error("No handler found for Redis channel",
				"channel", redisMsg.Channel,
				"pattern", redisMsg.Pattern)

			return fmt.Errorf("no handler found for channel: %s", redisMsg.Channel)
		}
	}

	s.logger.Debug("Parsing Redis message", "channel", redisMsg.Channel)

	message, err := FromJSON([]byte(redisMsg.Payload))
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	s.logger.Debug("Calling message handler",
		"channel", redisMsg.Channel,
		"message_id", message.GetMessageID())

	if handlerErr := handler(ctx, message); handlerErr != nil {
		return fmt.Errorf("handler failed for message %s: %w", message.GetMessageID(), handlerErr)
	}

	s.logger.Debug("Successfully processed Redis message",
		"channel", redisMsg.Channel,
		"message_id", message.GetMessageID())

	return nil
}

// SSubscribe subscribes to one or more sharded channels (Redis 7.0+).
// Creates a dedicated PubSub connection for each channel to ensure reliable message delivery.
// Sharded subscriptions use slot-based distribution for better scalability in Redis Cluster.
func (s *subscriber) SSubscribe(
	ctx context.Context,
	handler MessageHandler,
	channels ...string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, channel := range channels {
		// Check if already subscribed to this channel
		if s.shardedRunning[channel] {
			s.logger.Warn("Already subscribed to sharded channel, skipping", "channel", channel)
			continue
		}

		// Store handler for this channel
		s.handlers[channel] = handler

		// Create dedicated pubsub connection for this channel
		pubsub := s.client.SSubscribe(ctx, channel)
		s.shardedPubsubs[channel] = pubsub
		s.shardedRunning[channel] = true

		s.logger.Info("Created dedicated sharded subscription",
			"channel", channel)

		// Start dedicated message processor for this channel
		go s.processShardedMessages(ctx, channel, pubsub, handler)
	}

	s.logger.Info("Subscribed to sharded channels",
		"channels", channels,
		"total_sharded_subscriptions", len(s.shardedPubsubs))

	return nil
}

// processShardedMessages processes messages for a dedicated sharded channel subscription.
func (s *subscriber) processShardedMessages(
	ctx context.Context,
	channel string,
	pubsub *redis.PubSub,
	handler MessageHandler,
) {
	defer func() {
		s.mu.Lock()
		s.shardedRunning[channel] = false
		s.mu.Unlock()

		s.logger.Info("Sharded channel message processing stopped", "channel", channel)
	}()

	ch := pubsub.Channel(redis.WithChannelSize(s.config.ChannelBufferSize))

	s.logger.Info("Sharded channel message processing started",
		"channel", channel,
		"buffer_size", s.config.ChannelBufferSize)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info(
				"Context cancelled, stopping sharded message processing",
				"channel",
				channel,
			)
			return
		case msg, ok := <-ch:
			if !ok {
				s.logger.Warn("Sharded pubsub channel closed unexpectedly",
					"channel", channel)

				return
			}

			s.logger.Debug("Received message from Redis",
				"channel", msg.Channel,
				"pattern", msg.Pattern,
				"payload_size", len(msg.Payload))

			// Process message directly with the dedicated handler
			if err := s.handleShardedMessage(ctx, msg, handler); err != nil {
				s.logger.Error("Failed to handle sharded message",
					"channel", msg.Channel,
					"error", err)
			}
		}
	}
}

// handleShardedMessage processes a single Redis sharded message.
func (s *subscriber) handleShardedMessage(
	ctx context.Context,
	redisMsg *redis.Message,
	handler MessageHandler,
) error {
	s.logger.Debug("Parsing Redis message", "channel", redisMsg.Channel)

	message, err := FromJSON([]byte(redisMsg.Payload))
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	s.logger.Debug("Calling message handler",
		"channel", redisMsg.Channel,
		"message_id", message.GetMessageID())

	if handlerErr := handler(ctx, message); handlerErr != nil {
		return fmt.Errorf("handler failed for message %s: %w", message.GetMessageID(), handlerErr)
	}

	s.logger.Debug("Successfully processed Redis message",
		"channel", redisMsg.Channel,
		"message_id", message.GetMessageID())

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

	// If no more channels subscribed, close the connection to prevent stale state
	// This ensures fresh connections when resubscribing after all users disconnect
	if len(s.handlers) == 0 {
		s.logger.Info("No more active subscriptions, closing pubsub connection")

		if closeErr := s.pubsub.Close(); closeErr != nil {
			s.logger.Error("Failed to close pubsub connection", "error", closeErr)
		}

		s.pubsub = nil
		s.running = false

		s.logger.Info("Pubsub connection closed, ready for fresh subscription")
	}

	return nil
}

// SUnsubscribe unsubscribes from specified sharded channels.
func (s *subscriber) SUnsubscribe(channels ...string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, channel := range channels {
		// Get the dedicated pubsub for this channel
		pubsub, exists := s.shardedPubsubs[channel]
		if !exists {
			s.logger.Warn("Not subscribed to sharded channel", "channel", channel)
			continue
		}

		// Unsubscribe from the channel
		if err := pubsub.SUnsubscribe(context.Background(), channel); err != nil {
			s.logger.Error("Failed to unsubscribe from sharded channel",
				"channel", channel,
				"error", err)
			// Continue to clean up even on error
		}

		// Close the dedicated pubsub connection
		if err := pubsub.Close(); err != nil {
			s.logger.Error("Failed to close sharded pubsub connection",
				"channel", channel,
				"error", err)
		}

		// Remove from tracking maps
		delete(s.shardedPubsubs, channel)
		delete(s.shardedRunning, channel)
		delete(s.handlers, channel)

		s.logger.Info("Unsubscribed from sharded channel", "channel", channel)
	}

	s.logger.Info("Unsubscribed from sharded channels",
		"channels", channels,
		"remaining_sharded_subscriptions", len(s.shardedPubsubs))

	return nil
}

// Close closes the subscriber and releases resources.
func (s *subscriber) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error

	// Close regular pubsub connection
	if s.pubsub != nil {
		if err := s.pubsub.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close pubsub: %w", err))
		}

		s.pubsub = nil
	}

	// Close all sharded pubsub connections
	for channel, pubsub := range s.shardedPubsubs {
		if err := pubsub.Close(); err != nil {
			errs = append(
				errs,
				fmt.Errorf("failed to close sharded pubsub for %s: %w", channel, err),
			)
		}
	}

	// Clear all tracking maps
	s.handlers = make(map[string]MessageHandler)
	s.shardedPubsubs = make(map[string]*redis.PubSub)
	s.shardedRunning = make(map[string]bool)

	s.running = false
	s.logger.Info("Subscriber closed",
		"sharded_connections_closed", len(s.shardedPubsubs))

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}
