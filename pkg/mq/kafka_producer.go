package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// KafkaProducer is an interface for sending events to Kafka.
type KafkaProducer interface {
	Send(ctx context.Context, event BaseEvent) error
	Topic() string
}

// KafkaProducerConfig holds the configuration for the Kafka producer.
type KafkaProducerConfig struct {
	Brokers        []string
	ReturnSuccess  bool
	ReturnErrors   bool
	RetryMax       int
	FlushFrequency int // in milliseconds
	Acks           sarama.RequiredAcks
}

// KafkaSyncProducer implements the EventProducer interface using Kafka.
type KafkaSyncProducer struct {
	syncProducer sarama.SyncProducer
}

// KafkaAsyncProducer implements the EventProducer interface using Kafka.
type KafkaAsyncProducer struct {
	asyncProducer sarama.AsyncProducer
}

// BaseEvent represents a base event interface.
type BaseEvent interface {
	GetMetadata() KafkaMetadata
	GetPayload() interface{}
}

// KafkaMetadata provides common event properties.
type KafkaMetadata struct {
	EventID     uuid.UUID `json:"event_id"`
	EventType   string    `json:"event_type"`
	AggregateID uuid.UUID `json:"aggregate_id"`
	OccurredAt  time.Time `json:"occurred_at"`
	Source      string    `json:"source,omitempty"` // Service that produced the event
}

// KafkaAdminConfig holds the configuration for the Kafka admin.
type KafkaAdminConfig struct {
	Brokers []string
}

// KafkaAdmin represents a Kafka admin client.
type KafkaAdmin struct {
	Client sarama.Client
}

// NewKafkaAdmin creates a new instance of KafkaAdmin.
func NewKafkaAdmin(opt *KafkaAdminConfig) *KafkaAdmin {
	client, err := sarama.NewClient(opt.Brokers, sarama.NewConfig())
	if err != nil {
		log.Fatalf("failed to create kafka admin: %v", err)
	}

	return &KafkaAdmin{
		Client: client,
	}
}

// CreateTopic creates a new Kafka topic.
func (admin *KafkaAdmin) CreateTopic(topic string, numPartitions, replicationFactor int) {
	adminClient, err := sarama.NewClusterAdminFromClient(admin.Client)
	if err != nil {
		log.Fatalf("failed to create kafka admin: %v", err)
	}

	topics, err := adminClient.ListTopics()
	if err != nil {
		log.Fatalf("failed to list topics: %v", err)
	}

	if _, exists := topics[topic]; !exists {
		topicDetail := &sarama.TopicDetail{
			NumPartitions:     int32(numPartitions),
			ReplicationFactor: int16(replicationFactor),
		}
		if err := adminClient.CreateTopic(topic, topicDetail, false); err != nil {
			log.Fatalf("failed to create topic: %v", err)
		}
	}
}

// NewKafkaSyncProducer creates a new instance of sync ProducerKafka.
func NewKafkaSyncProducer(cfg *KafkaProducerConfig) (*KafkaSyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = cfg.Acks
	config.Producer.Return.Successes = cfg.ReturnSuccess
	config.Producer.Return.Errors = cfg.ReturnErrors
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = cfg.RetryMax
	config.Producer.Retry.Backoff = time.Millisecond * time.Duration(cfg.FlushFrequency)

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaSyncProducer{
		syncProducer: producer,
	}, nil
}

// NewKafkaAsyncProducer creates a new instance of async KafkaAsyncProducer.
func NewKafkaAsyncProducer(cfg *KafkaProducerConfig) (*KafkaAsyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = cfg.Acks
	config.Producer.Return.Successes = cfg.ReturnSuccess
	config.Producer.Return.Errors = cfg.ReturnErrors
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = cfg.RetryMax
	config.Producer.Retry.Backoff = time.Millisecond * time.Duration(cfg.FlushFrequency)

	producer, err := sarama.NewAsyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka async producer: %w", err)
	}

	return &KafkaAsyncProducer{
		asyncProducer: producer,
	}, nil
}

// ProduceSync an event to Kafka.
func (p *KafkaSyncProducer) ProduceSync(topic string, evt BaseEvent) error {
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
	partition, offset, err := p.syncProducer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send sync message to Kafka: %w", err)
	}

	log.Printf("Event published to Kafka - Topic: %s, Partition: %d, Offset: %d, Type: %s",
		topic, partition, offset, metadata.EventType)

	return nil
}

// CloseSync closes the Kafka producer.
func (p *KafkaSyncProducer) CloseSync() error {
	if p.syncProducer != nil {
		log.Println("Closing Kafka sync producer")

		return p.syncProducer.Close()
	}

	return nil
}

// ProduceAsync an event to Kafka.
func (p *KafkaAsyncProducer) ProduceAsync(ctx context.Context, topic string, evt BaseEvent) error {
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

	// Send message asynchronously
	select {
	case p.asyncProducer.Input() <- message:
		log.Printf("Event sent to async producer - Topic: %s, Type: %s",
			topic, metadata.EventType)

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// CloseAsync closes the Kafka async producer (FIXED: now uses correct receiver).
func (p *KafkaAsyncProducer) CloseAsync() error {
	if p.asyncProducer != nil {
		log.Println("Closing Kafka async producer")

		return p.asyncProducer.Close()
	}

	return nil
}
