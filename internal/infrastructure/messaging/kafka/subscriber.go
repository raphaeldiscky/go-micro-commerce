package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"

	"github.com/raphaeldiscky/go-ddd-template/internal/domain/events"
)

// KafkaEventSubscriber implements the EventSubscriber interface using Kafka.
type KafkaEventSubscriber struct {
	consumer      sarama.ConsumerGroup
	handlers      map[string]events.EventHandler
	handlersMutex sync.RWMutex
	topics        []string
	groupID       string
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewKafkaEventSubscriber creates a new Kafka event subscriber.
func NewKafkaEventSubscriber(
	brokers []string,
	groupID string,
	topics []string,
) (*KafkaEventSubscriber, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Group.Session.Timeout = sarama.NewConfig().Consumer.Group.Session.Timeout
	config.Consumer.Group.Heartbeat.Interval = sarama.NewConfig().Consumer.Group.Heartbeat.Interval

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	return &KafkaEventSubscriber{
		consumer: consumer,
		handlers: make(map[string]events.EventHandler),
		topics:   topics,
		groupID:  groupID,
	}, nil
}

// Subscribe subscribes to an event type with a handler.
func (s *KafkaEventSubscriber) Subscribe(
	ctx context.Context,
	eventType string,
	handler events.EventHandler,
) error {
	s.handlersMutex.Lock()
	s.handlers[eventType] = handler
	s.handlersMutex.Unlock()

	log.Printf("Subscribed to event type: %s", eventType)

	return nil
}

// Unsubscribe removes the handler for an event type.
func (s *KafkaEventSubscriber) Unsubscribe(ctx context.Context, eventType string) error {
	s.handlersMutex.Lock()
	delete(s.handlers, eventType)
	s.handlersMutex.Unlock()

	log.Printf("Unsubscribed from event type: %s", eventType)

	return nil
}

// Start starts consuming messages from Kafka.
func (s *KafkaEventSubscriber) Start(ctx context.Context) error {
	consumerCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-consumerCtx.Done():
				log.Println("Kafka consumer context canceled")

				return
			default:
				if err := s.consumer.Consume(consumerCtx, s.topics, s); err != nil {
					log.Printf("Error consuming from Kafka: %v", err)

					return
				}
			}
		}
	}()

	log.Printf("Started Kafka consumer for group: %s, topics: %v", s.groupID, s.topics)

	return nil
}

// Stop stops the Kafka consumer.
func (s *KafkaEventSubscriber) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}

	s.wg.Wait()

	return s.consumer.Close()
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (s *KafkaEventSubscriber) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Kafka consumer group session setup")

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (s *KafkaEventSubscriber) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Kafka consumer group session cleanup")

	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (s *KafkaEventSubscriber) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			if err := s.handleMessage(session.Context(), message); err != nil {
				log.Printf("Error handling message: %v", err)
				// In production, you might want to send to dead letter queue
				continue
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleMessage processes a Kafka message.
func (s *KafkaEventSubscriber) handleMessage(
	ctx context.Context,
	message *sarama.ConsumerMessage,
) error {
	// Extract event type from headers
	var eventType string

	for _, header := range message.Headers {
		if string(header.Key) == "event_type" {
			eventType = string(header.Value)

			break
		}
	}

	if eventType == "" {
		return fmt.Errorf("message missing event_type header")
	}

	s.handlersMutex.RLock()
	handler, exists := s.handlers[eventType]
	s.handlersMutex.RUnlock()

	if !exists {
		log.Printf("No handler registered for event type: %s", eventType)

		return nil
	}

	// Unmarshal the event
	var baseEvent events.BaseDomainEvent
	if err := json.Unmarshal(message.Value, &baseEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Call the handler
	if err := handler(ctx, baseEvent); err != nil {
		return fmt.Errorf("handler failed for event %s: %w", baseEvent.EventID(), err)
	}

	log.Printf("Successfully processed event: %s (type: %s)", baseEvent.EventID(), eventType)

	return nil
}
