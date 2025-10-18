package redis

import (
	"context"
	"fmt"
	"time"
)

// publisher implements the Publisher interface.
type publisher struct {
	client PubSubClient
	config PubSubConfig
}

// NewPublisher creates a new Redis publisher.
func NewPublisher(client PubSubClient, config PubSubConfig) Publisher {
	return &publisher{
		client: client,
		config: config,
	}
}

// Publish publishes a message to the specified channel.
func (p *publisher) Publish(ctx context.Context, channel string, message *Message) error {
	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	err = p.client.Publish(ctx, channel, data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish message to channel %s: %w", channel, err)
	}

	return nil
}

// PublishWithRetry publishes a message with retry logic.
func (p *publisher) PublishWithRetry(
	ctx context.Context,
	channel string,
	message *Message,
	maxRetries int,
) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := time.Duration(attempt) * p.config.RetryDelay
			if delay > p.config.MaxRetryDelay {
				delay = p.config.MaxRetryDelay
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := p.Publish(ctx, channel, message)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return fmt.Errorf("failed to publish after %d attempts: %w", maxRetries+1, lastErr)
}

// SPublish publishes a message to a sharded channel (Redis 7.0+).
// Sharded pub/sub uses slot-based distribution for better scalability in Redis Cluster.
func (p *publisher) SPublish(ctx context.Context, channel string, message *Message) error {
	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	err = p.client.SPublish(ctx, channel, data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish message to sharded channel %s: %w", channel, err)
	}

	return nil
}

// SPublishWithRetry publishes a message to a sharded channel with retry logic.
func (p *publisher) SPublishWithRetry(
	ctx context.Context,
	channel string,
	message *Message,
	maxRetries int,
) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := time.Duration(attempt) * p.config.RetryDelay
			if delay > p.config.MaxRetryDelay {
				delay = p.config.MaxRetryDelay
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := p.SPublish(ctx, channel, message)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	return fmt.Errorf(
		"failed to publish to sharded channel after %d attempts: %w",
		maxRetries+1,
		lastErr,
	)
}

// Close closes the publisher and releases resources.
func (p *publisher) Close() error {
	// Redis client is shared, so we don't close it here
	// The application should manage the Redis client lifecycle
	return nil
}
