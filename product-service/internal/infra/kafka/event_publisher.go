// Package kafka provides an implementation of the EventPublisher interface using Kafka.
package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/event"
)

// EventPublisherKafka implements the EventPublisher interface using Kafka.
type EventPublisherKafka struct {
	producer sarama.SyncProducer
	topic    string
}

// NewEventPublisherKafka creates a new instance of EventPublisherKafka.
func NewEventPublisherKafka(brokers []string, topic string) (*EventPublisherKafka, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Retry.Backoff = 100 * time.Millisecond

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &EventPublisherKafka{
		producer: producer,
		topic:    topic,
	}, nil
}

// Publish publishes an event to Kafka.
func (p *EventPublisherKafka) Publish(evt event.DomainEvent) error {
	// Marshal event to JSON
	eventData, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	message := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(evt.GetAggregateID().String()),
		Value: sarama.ByteEncoder(eventData),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event-type"),
				Value: []byte(evt.GetEventType()),
			},
			{
				Key:   []byte("timestamp"),
				Value: []byte(evt.GetOccurredAt().Format(time.RFC3339)),
			},
		},
	}

	// Send message
	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	log.Printf("Event published to Kafka - Topic: %s, Partition: %d, Offset: %d, Type: %s",
		p.topic, partition, offset, evt.GetEventType())

	return nil
}

// Close closes the Kafka producer.
func (p *EventPublisherKafka) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}

	return nil
}
