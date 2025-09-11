// Package kafka provides a Kafka consumer implementation for consuming messages from Kafka topics.
package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// Handler is a function type for handling Kafka messages.
// It receives the context and the message body as bytes.
type Handler func(ctx context.Context, body []byte) error

// Consumer is the interface that all consumers must implement.
type Consumer interface {
	Consume(ctx context.Context) error
	Close() error
	Topic() string
}

// consumerKafka handles consuming events from Kafka. It implements the sarama.ConsumerGroupHandler.
type consumerKafka struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	handler       Handler // The business logic handler for a message.
	appLogger     logger.Logger
}

// NewConsumer creates and configures a new Kafka consumer.
func NewConsumer(
	brokers []string,
	topic, groupID string,
	handler Handler,
	appLogger logger.Logger,
) (Consumer, error) {
	if err := TestConnection(brokers, appLogger); err != nil {
		return nil, fmt.Errorf("kafka connection test failed: %w", err)
	}

	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0 // Or your desired Kafka version
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true

	// Create a consumer group
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	appLogger.Infof("successfully created consumer for topic: %s, group: %s", topic, groupID)

	return &consumerKafka{
		consumerGroup: consumerGroup,
		topic:         topic,
		handler:       handler,
		appLogger:     appLogger,
	}, nil
}

// TestConnection tests the connection to the Kafka brokers.
func TestConnection(brokers []string, appLogger logger.Logger) error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0

	// Create a simple client to test connection
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return fmt.Errorf("failed to create Kafka client: %w", err)
	}

	defer func() {
		if err = client.Close(); err != nil {
			appLogger.Errorf("Error closing Kafka client: %v", err)
		}
	}()

	// Test if we can get broker list
	brokersList := client.Brokers()
	if len(brokersList) == 0 {
		return errors.New("no brokers available")
	}

	appLogger.Infof("successfully connected to Kafka. Available brokers: %d", len(brokersList))

	return nil
}

// Consume starts the consumer group and listens for messages. This is a blocking call.
func (c *consumerKafka) Consume(ctx context.Context) error {
	c.appLogger.Infof("Consumer for topic '%s' starting...", c.topic)

	for {
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will end
		// and `Consume` will return.
		if err := c.consumerGroup.Consume(ctx, []string{c.topic}, c); err != nil {
			c.appLogger.Errorf("Error from consumer for topic %s: %v", c.topic, err)
			// If context is canceled, we should stop.
			if ctx.Err() != nil {
				return err
			}
		}

		// Check if context was canceled, signaling that we should stop.
		if ctx.Err() != nil {
			c.appLogger.Infof("Context canceled for topic %s. Exiting consumer loop.", c.topic)

			return ctx.Err()
		}
	}
}

// Topic returns the consumer's topic.
func (c *consumerKafka) Topic() string {
	return c.topic
}

// Close shuts down the consumer group.
func (c *consumerKafka) Close() error {
	c.appLogger.Infof("Closing consumer for topic: %s", c.topic)

	if c.consumerGroup != nil {
		return c.consumerGroup.Close()
	}

	return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *consumerKafka) Setup(sarama.ConsumerGroupSession) error {
	c.appLogger.Infof("Consumer group session started for topic: %s", c.topic)

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (c *consumerKafka) Cleanup(sarama.ConsumerGroupSession) error {
	c.appLogger.Infof("Consumer group session ended for topic: %s", c.topic)

	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *consumerKafka) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	// The `ConsumeClaim` function is called for each claimed partition.
	// It must run in a loop to process messages from the claim's Messages channel.
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				c.appLogger.Infof("Message channel closed for topic: %s", c.topic)

				return nil
			}

			// Call the specific handler with the message value.
			err := c.handler(session.Context(), message.Value)
			if err != nil {
				c.appLogger.Errorf("Handler error for topic %s: %v", c.topic, err)
				// TODO: Implement retry logic or dead letter queue
				// For now, we'll still mark the message to avoid infinite reprocessing
				// In production, you might want to:
				// - Retry with exponential backoff
				// - Send to dead letter queue after max retries
				// - Or don't mark the message (but this can cause infinite loops)
			}

			// Only mark message as processed after handling (success or failure)
			session.MarkMessage(message, "")

		// Should return when the session is canceled, which happens on a rebalance.
		case <-session.Context().Done():
			c.appLogger.Infof("Session context canceled for topic: %s", c.topic)

			return nil
		}
	}
}
