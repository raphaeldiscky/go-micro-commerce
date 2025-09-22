package redis

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// PubSubManager manages both publisher and subscriber instances using an existing Redis client.
type PubSubManager struct {
	client    PubSubClient
	config    PubSubConfig
	logger    logger.Logger
	publisher Publisher
}

// NewPubSubManager creates a new pub/sub manager with an existing Redis client.
func NewPubSubManager(
	client PubSubClient,
	config PubSubConfig,
	logger logger.Logger,
) *PubSubManager {
	return &PubSubManager{
		client:    client,
		config:    config,
		logger:    logger,
		publisher: NewPublisher(client, config),
	}
}

// GetPublisher returns the publisher instance.
func (m *PubSubManager) GetPublisher() Publisher {
	return m.publisher
}

// CreateSubscriber creates a new subscriber instance.
func (m *PubSubManager) CreateSubscriber() Subscriber {
	return NewSubscriber(m.client, m.config, m.logger)
}

// GetClient returns the underlying Redis client.
func (m *PubSubManager) GetClient() PubSubClient {
	return m.client
}

// Close closes the pub/sub manager and releases resources.
// Note: This does not close the Redis client as it may be shared.
func (m *PubSubManager) Close() error {
	if err := m.publisher.Close(); err != nil {
		m.logger.Errorf("Failed to close publisher: %v", err)
		return err
	}

	return nil
}
