package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/raphaeldiscky/go-ddd-template/internal/domain/events"
)

// KafkaEventPublisher implements the EventPublisher interface using Kafka
type KafkaEventPublisher struct {
	producer      sarama.SyncProducer
	topicPrefix   string
	retryAttempts int
	retryDelay    time.Duration
}

// KafkaConfig holds configuration for Kafka
type KafkaConfig struct {
	Brokers       []string
	TopicPrefix   string
	RetryAttempts int
	RetryDelay    time.Duration
}

// NewKafkaEventPublisher creates a new Kafka event publisher
func NewKafkaEventPublisher(config KafkaConfig) (*KafkaEventPublisher, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = config.RetryAttempts
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Compression = sarama.CompressionSnappy
	saramaConfig.Producer.Idempotent = true
	saramaConfig.Net.MaxOpenRequests = 1

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaEventPublisher{
		producer:      producer,
		topicPrefix:   config.TopicPrefix,
		retryAttempts: config.RetryAttempts,
		retryDelay:    config.RetryDelay,
	}, nil
}

// Publish publishes a single domain event to Kafka
func (p *KafkaEventPublisher) Publish(ctx context.Context, event events.DomainEvent) error {
	return p.publishWithRetry(ctx, event)
}

// PublishBatch publishes multiple domain events to Kafka
func (p *KafkaEventPublisher) PublishBatch(ctx context.Context, eventList []events.DomainEvent) error {
	for _, event := range eventList {
		if err := p.publishWithRetry(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.EventID(), err)
		}
	}
	return nil
}

// publishWithRetry publishes an event with retry logic
func (p *KafkaEventPublisher) publishWithRetry(ctx context.Context, event events.DomainEvent) error {
	var lastErr error

	for attempt := 0; attempt <= p.retryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(p.retryDelay):
			}
		}

		if err := p.publishEvent(ctx, event); err != nil {
			lastErr = err
			log.Printf("Failed to publish event (attempt %d/%d): %v", attempt+1, p.retryAttempts+1, err)
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to publish event after %d attempts: %w", p.retryAttempts+1, lastErr)
}

// publishEvent publishes a single event to Kafka
func (p *KafkaEventPublisher) publishEvent(ctx context.Context, event events.DomainEvent) error {
	topic := p.getTopicName(event.EventType())

	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(event.AggregateID()),
		Value: sarama.ByteEncoder(eventData),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_id"),
				Value: []byte(event.EventID()),
			},
			{
				Key:   []byte("event_type"),
				Value: []byte(event.EventType()),
			},
			{
				Key:   []byte("aggregate_type"),
				Value: []byte(event.AggregateType()),
			},
			{
				Key:   []byte("occurred_at"),
				Value: []byte(event.OccurredAt().Format(time.RFC3339)),
			},
		},
	}

	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	log.Printf("Event published to Kafka: topic=%s, partition=%d, offset=%d, event_id=%s",
		topic, partition, offset, event.EventID())

	return nil
}

// getTopicName generates the topic name for an event type
func (p *KafkaEventPublisher) getTopicName(eventType string) string {
	if p.topicPrefix != "" {
		return fmt.Sprintf("%s.%s", p.topicPrefix, eventType)
	}
	return eventType
}

// Close closes the Kafka producer
func (p *KafkaEventPublisher) Close() error {
	return p.producer.Close()
}
