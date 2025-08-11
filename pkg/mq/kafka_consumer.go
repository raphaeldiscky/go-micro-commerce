// Package mq provides a Kafka consumer implementation for consuming messages from Kafka topics.
package mq

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

// KafkaHandler is a function type for handling Kafka messages.
// It receives the context and the message body as bytes.
type KafkaHandler func(ctx context.Context, body []byte) error

// KafkaConsumer is the interface that all consumers must implement.
type KafkaConsumer interface {
	Consume(ctx context.Context) error
	Close() error
	Topic() string
}

// consumerKafka handles consuming events from Kafka. It implements the sarama.ConsumerGroupHandler.
type consumerKafka struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	handler       KafkaHandler // The business logic handler for a message.
}

// NewConsumerKafka creates and configures a new Kafka consumer.
func NewConsumerKafka(
	brokers []string,
	topic, groupID string,
	handler KafkaHandler,
) (KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0 // Or your desired Kafka version
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true

	// Create a consumer group
	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &consumerKafka{
		consumerGroup: consumerGroup,
		topic:         topic,
		handler:       handler,
	}, nil
}

// Consume starts the consumer group and listens for messages. This is a blocking call.
func (c *consumerKafka) Consume(ctx context.Context) error {
	log.Printf("Consumer for topic '%s' starting...", c.topic)

	for {
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will end
		// and `Consume` will return.
		if err := c.consumerGroup.Consume(ctx, []string{c.topic}, c); err != nil {
			log.Printf("Error from consumer for topic %s: %v", c.topic, err)
			// If context is canceled, we should stop.
			if ctx.Err() != nil {
				return err
			}
		}

		// Check if context was canceled, signaling that we should stop.
		if ctx.Err() != nil {
			log.Printf("Context canceled for topic %s. Exiting consumer loop.", c.topic)

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
	log.Printf("Closing consumer for topic: %s", c.topic)

	if c.consumerGroup != nil {
		return c.consumerGroup.Close()
	}

	return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *consumerKafka) Setup(sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group session started for topic: %s", c.topic)

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (c *consumerKafka) Cleanup(sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group session ended for topic: %s", c.topic)

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
				log.Printf("Message channel closed for topic: %s", c.topic)

				return nil
			}

			// Call the specific handler with the message value.
			err := c.handler(session.Context(), message.Value)
			if err != nil {
				log.Printf("Handler error for topic %s: %v", c.topic, err)
				// Here you could implement retry logic, or send to a dead-letter queue.
				// For now, we just log and continue.
			}

			// Mark the message as processed.
			session.MarkMessage(message, "")

		// Should return when the session is canceled, which happens on a rebalance.
		case <-session.Context().Done():
			return nil
		}
	}
}
