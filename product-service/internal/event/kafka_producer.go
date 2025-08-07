package event

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
)

// ProducerKafka implements the EventProducer interface using Kafka.
type ProducerKafka struct {
	producer sarama.SyncProducer
}

// NewProducerKafka creates a new instance of ProducerKafka.
func NewProducerKafka(cfg *config.KafkaConfig) (*ProducerKafka, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = cfg.ReturnSuccess
	config.Producer.Return.Errors = cfg.ReturnErrors
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = cfg.RetryMax
	config.Producer.Retry.Backoff = time.Millisecond * time.Duration(cfg.FlushFrequency)

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &ProducerKafka{
		producer: producer,
	}, nil
}

// Publish publishes an event to Kafka.
func (p *ProducerKafka) Publish(topic string, evt BaseEvent) error {
	// Marshal event to JSON
	eventData, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	metadata := evt.GetMetadata()

	// Create Kafka message
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(metadata.AggregateID.String()),
		Value: sarama.ByteEncoder(eventData),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event-type"),
				Value: []byte(metadata.EventType),
			},
			{
				Key:   []byte("timestamp"),
				Value: []byte(metadata.OccurredAt.Format(time.RFC3339)),
			},
		},
	}

	// Send message
	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	log.Printf("Event published to Kafka - Topic: %s, Partition: %d, Offset: %d, Type: %s",
		topic, partition, offset, metadata.EventType)

	return nil
}

// Close closes the Kafka producer.
func (p *ProducerKafka) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}

	return nil
}
