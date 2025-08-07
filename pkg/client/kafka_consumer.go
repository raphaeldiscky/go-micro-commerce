package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// BaseEvent represents a base event interface.
type BaseEvent interface {
	GetMetadata() KafkaMetadata
	GetPayload() interface{}
}

// KafkaConsumerConfig holds the configuration for the Kafka consumer.
type KafkaConsumerConfig struct {
	Brokers        []string
	GroupID        string
	Topics         []string
	ReturnSuccess  bool
	ReturnErrors   bool
	RetryMax       int
	FlushFrequency int // in milliseconds
}

// ConsumerHandler defines the handler function for consuming messages.
type ConsumerHandler func(message *BaseEvent) error

// ConsumerKafka handles consuming events from Kafka.
type ConsumerKafka struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	handler       ConsumerHandler
}

// NewConsumerKafka creates a new Kafka consumer group.
func NewConsumerKafka(ctx context.Context, cfg *KafkaConsumerConfig, groupID, topic string, handler ConsumerHandler) (*ConsumerKafka, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Version = sarama.V2_6_0_0
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaCfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, groupID, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	return &ConsumerKafka{
		consumerGroup: consumerGroup,
		topic:         topic,
		handler:       handler,
	}, nil
}

// Start begins consuming messages in a blocking loop.
func (c *ConsumerKafka) Start(ctx context.Context) error {
	for {
		if err := c.consumerGroup.Consume(ctx, []string{c.topic}, c); err != nil {
			log.Printf("Error consuming from topic %s: %v", c.topic, err)
			continue
		}

		// Check if context was canceled
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// Close shuts down the consumer group.
func (c *ConsumerKafka) Close() error {
	if c.consumerGroup != nil {
		return c.consumerGroup.Close()
	}
	return nil
}

// Setup is called before consuming messages.
func (c *ConsumerKafka) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup is called after consuming messages.
func (c *ConsumerKafka) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim processes messages from the claim.
func (c *ConsumerKafka) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var evt BaseEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		if err := c.handler(&evt); err != nil {
			log.Printf("Handler error: %v", err)
			continue
		}

		session.MarkMessage(msg, "")
	}

	return nil
}
