// Package redis provides universal Redis pub/sub functionality for microservices.
package redis

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Publisher defines the interface for publishing messages to Redis channels.
type Publisher interface {
	// Publish publishes a message to the specified channel.
	Publish(ctx context.Context, channel string, message *Message) error
	// PublishWithRetry publishes a message with retry logic.
	PublishWithRetry(ctx context.Context, channel string, message *Message, maxRetries int) error
	// Close closes the publisher and releases resources.
	Close() error
}

// Subscriber defines the interface for subscribing to Redis channels.
type Subscriber interface {
	// Subscribe subscribes to one or more channels and calls the handler for each message.
	Subscribe(ctx context.Context, handler MessageHandler, channels ...string) error
	// SubscribePattern subscribes to channels matching a pattern.
	SubscribePattern(ctx context.Context, handler MessageHandler, pattern string) error
	// Unsubscribe unsubscribes from specified channels.
	Unsubscribe(channels ...string) error
	// Close closes the subscriber and releases resources.
	Close() error
}

// MessageHandler is a function type for handling received messages.
type MessageHandler func(ctx context.Context, message *Message) error

// PubSubConfig holds configuration for Redis pub/sub.
type PubSubConfig struct {
	// RetryAttempts is the number of retry attempts for failed operations.
	RetryAttempts int `mapstructure:"retry_attempts"`
	// RetryDelay is the initial delay between retries.
	RetryDelay time.Duration `mapstructure:"retry_delay"`
	// MaxRetryDelay is the maximum delay between retries.
	MaxRetryDelay time.Duration `mapstructure:"max_retry_delay"`
	// ChannelBufferSize is the buffer size for subscription channels.
	ChannelBufferSize int `mapstructure:"channel_buffer_size"`
}

// DefaultPubSubConfig returns the default pub/sub configuration.
func DefaultPubSubConfig() PubSubConfig {
	return PubSubConfig{
		RetryAttempts:     DefaultRetryAttempts,
		RetryDelay:        DefaultRetryDelayMs * time.Millisecond,
		MaxRetryDelay:     DefaultMaxRetryDelaySec * time.Second,
		ChannelBufferSize: DefaultChannelBufferSize,
	}
}

// MessageMetadata contains metadata for pub/sub messages.
type MessageMetadata struct {
	// MessageID is a unique identifier for the message.
	MessageID string `json:"message_id"`
	// CorrelationID is used for distributed tracing.
	CorrelationID string `json:"correlation_id,omitempty"`
	// Source is the service that published the message.
	Source string `json:"source"`
	// Timestamp is when the message was created.
	Timestamp time.Time `json:"timestamp"`
	// ContentType describes the message payload format.
	ContentType string `json:"content_type"`
	// Version is the message schema version.
	Version string `json:"version,omitempty"`
}

// NewMessageMetadata creates a new message metadata with default values.
func NewMessageMetadata(source string) MessageMetadata {
	return MessageMetadata{
		MessageID:   uuid.New().String(),
		Source:      source,
		Timestamp:   time.Now(),
		ContentType: "application/json",
	}
}

// SetCorrelationID sets the correlation ID for distributed tracing.
func (m *MessageMetadata) SetCorrelationID(correlationID string) {
	m.CorrelationID = correlationID
}
