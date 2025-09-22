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
	client  *redis.Client
	config  PubSubConfig
	logger  logger.Logger
	pubsub  *redis.PubSub
	mu      sync.RWMutex
	running bool
}

// NewSubscriber creates a new Redis subscriber.
func NewSubscriber(client *redis.Client, config PubSubConfig, logger logger.Logger) Subscriber {
	return &subscriber{
		client: client,
		config: config,
		logger: logger,
	}
}

// Subscribe subscribes to one or more channels and calls the handler for each message.
func (s *subscriber) Subscribe(
	ctx context.Context,
	handler MessageHandler,
	channels ...string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("subscriber is already running")
	}

	s.pubsub = s.client.Subscribe(ctx, channels...)
	s.running = true

	go s.processMessages(ctx, handler)

	s.logger.Infof("Subscribed to channels: %v", channels)

	return nil
}

// SubscribePattern subscribes to channels matching a pattern.
func (s *subscriber) SubscribePattern(
	ctx context.Context,
	handler MessageHandler,
	pattern string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("subscriber is already running")
	}

	s.pubsub = s.client.PSubscribe(ctx, pattern)
	s.running = true

	go s.processMessages(ctx, handler)

	s.logger.Infof("Subscribed to pattern: %s", pattern)

	return nil
}

// processMessages processes incoming messages in a separate goroutine.
func (s *subscriber) processMessages(ctx context.Context, handler MessageHandler) {
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

			if err := s.handleMessage(ctx, msg, handler); err != nil {
				s.logger.Errorf("Failed to handle message: %v", err)
			}
		}
	}
}

// handleMessage processes a single Redis message.
func (s *subscriber) handleMessage(
	ctx context.Context,
	redisMsg *redis.Message,
	handler MessageHandler,
) error {
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.pubsub == nil {
		return errors.New("not subscribed to any channels")
	}

	err := s.pubsub.Unsubscribe(context.Background(), channels...)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from channels %v: %w", channels, err)
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

	s.running = false
	s.logger.Info("Subscriber closed")

	return nil
}
