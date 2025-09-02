package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// ProducerInterface is an interface for sending events to Kafka.
type ProducerInterface interface {
	Send(ctx context.Context, event BaseEvent) error
	Topic() string
}

// ProducerConfig holds the configuration for the Kafka producer.
type ProducerConfig struct {
	Brokers        []string
	ReturnSuccess  bool
	ReturnErrors   bool
	RetryMax       int
	FlushFrequency int // in milliseconds
	Acks           sarama.RequiredAcks
}

// SyncProducer implements the EventProducer interface using Kafka.
type SyncProducer struct {
	producer sarama.SyncProducer
}

// AsyncProducer implements the EventProducer interface using Kafka.
type AsyncProducer struct {
	producer  sarama.AsyncProducer
	RetryChan chan *sarama.ProducerMessage
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewSyncProducer creates a new instance of sync ProducerKafka.
func NewSyncProducer(cfg *ProducerConfig) (*SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = cfg.Acks
	config.Producer.Return.Successes = cfg.ReturnSuccess
	config.Producer.Return.Errors = cfg.ReturnErrors
	// Don't override the Acks setting
	if cfg.Acks == 0 {
		config.Producer.RequiredAcks = sarama.WaitForAll
	}

	config.Producer.Retry.Max = cfg.RetryMax
	config.Producer.Retry.Backoff = time.Millisecond * time.Duration(cfg.FlushFrequency)

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &SyncProducer{
		producer: producer,
	}, nil
}

// NewAsyncProducer creates a new instance of async AsyncProducer.
func NewAsyncProducer(cfg *ProducerConfig) (*AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = cfg.Acks
	config.Producer.Return.Successes = cfg.ReturnSuccess
	config.Producer.Return.Errors = cfg.ReturnErrors
	// Don't override the Acks setting
	if cfg.Acks == 0 {
		config.Producer.RequiredAcks = sarama.WaitForAll
	}

	config.Producer.Retry.Max = cfg.RetryMax
	config.Producer.Retry.Backoff = time.Millisecond * time.Duration(cfg.FlushFrequency)

	producer, err := sarama.NewAsyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka async producer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	asyncProducer := &AsyncProducer{
		producer:  producer,
		RetryChan: make(chan *sarama.ProducerMessage, 1000), // Buffered channel
		ctx:       ctx,
		cancel:    cancel,
	}
	// background goroutines
	asyncProducer.wg.Add(1)
	go asyncProducer.handleErrors()

	asyncProducer.wg.Add(1)
	go asyncProducer.handleRetries()

	asyncProducer.wg.Add(1)
	go asyncProducer.handleSuccesses()

	return asyncProducer, nil
}

// ProduceSync an event to Kafka.
func (p *SyncProducer) ProduceSync(topic string, evt BaseEvent) error {
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
		return fmt.Errorf("failed to send sync message to Kafka: %w", err)
	}

	log.Printf("Event published to Kafka - Topic: %s, Partition: %d, Offset: %d, Type: %s",
		topic, partition, offset, metadata.EventType)

	return nil
}

// CloseSync closes the Kafka producer.
func (p *SyncProducer) CloseSync() error {
	if p.producer != nil {
		log.Println("Closing Kafka sync producer")

		return p.producer.Close()
	}

	return nil
}

// ProduceAsync sends an event to Kafka asynchronously.
func (p *AsyncProducer) ProduceAsync(ctx context.Context, topic string, evt BaseEvent) error {
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
	case p.producer.Input() <- message:
		log.Printf("Event sent to async producer - Topic: %s, Type: %s",
			topic, metadata.EventType)

		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

// CloseAsync closes the Kafka async producer.
func (p *AsyncProducer) CloseAsync() error {
	if p.producer != nil {
		log.Println("Closing Kafka async producer")

		// Signal shutdown
		p.cancel()

		// Wait for goroutines to finish
		p.wg.Wait()

		// Close the retry channel
		close(p.RetryChan)

		return p.producer.Close()
	}

	return nil
}

// handleSuccesses handles successful message deliveries.
func (p *AsyncProducer) handleSuccesses() {
	defer p.wg.Done()

	for {
		select {
		case msg := <-p.producer.Successes():
			if msg != nil {
				log.Printf("Message delivered successfully: topic=%s, partition=%d, offset=%d",
					msg.Topic, msg.Partition, msg.Offset)
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// handleErrors handles errors that occur during message production.
func (p *AsyncProducer) handleErrors() {
	defer p.wg.Done()

	for {
		select {
		case err := <-p.producer.Errors():
			if err != nil {
				log.Printf("failed to send message: %v", err.Err)

				// Add to retry channel if there's space
				select {
				case p.RetryChan <- err.Msg:
					log.Printf("Message queued for retry: Topic=%s", err.Msg.Topic)
				default:
					log.Printf("Retry channel full, dropping message: Topic=%s", err.Msg.Topic)
				}
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// handleRetries processes messages from the retry channel.
func (p *AsyncProducer) handleRetries() {
	defer p.wg.Done()

	retryTicker := time.NewTicker(2 * time.Second) // Retry every 2 seconds
	defer retryTicker.Stop()

	for {
		select {
		case <-retryTicker.C:
			// Process all messages in retry channel
		inner:
			for {
				select {
				case msg := <-p.RetryChan:
					// Retry sending the message
					select {
					case p.producer.Input() <- msg:
						log.Printf("Message retried successfully: Topic=%s", msg.Topic)
					case <-p.ctx.Done():
						return
					default:
						// Producer input channel is full, put message back in retry channel
						select {
						case p.RetryChan <- msg:
							log.Printf("Producer busy, message re-queued for retry: Topic=%s", msg.Topic)
						default:
							log.Printf("Retry channel full, dropping retried message: Topic=%s", msg.Topic)
						}
					}
				default:
					// No more messages to retry
					break inner
				}
			}
		case <-p.ctx.Done():
			return
		}
	}
}
