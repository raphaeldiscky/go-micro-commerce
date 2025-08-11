package event

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/constant"
)

// UserVerifiedPayload holds the data for the user verified event.
type UserVerifiedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

// UserVerifiedEvent is the envelope for all user verified events.
type UserVerifiedEvent struct {
	metadata KafkaMetadata
	payload  UserVerifiedPayload
}

// GetPayload returns the data associated with the UserVerifiedEvent.
func (e *UserVerifiedEvent) GetPayload() interface{} {
	return e.payload
}

// GetMetadata returns the metadata associated with the UserVerifiedEvent.
func (e *UserVerifiedEvent) GetMetadata() KafkaMetadata {
	return e.metadata
}

// NewUserVerifiedEvent creates a new UserVerifiedEvent.
func NewUserVerifiedEvent(
	userID uuid.UUID,
	email string,
) *UserVerifiedEvent {
	return &UserVerifiedEvent{
		metadata: KafkaMetadata{
			EventID:     uuid.New(),
			EventType:   constant.KafkaEventTypeUserVerified,
			AggregateID: userID,
			OccurredAt:  time.Now().UTC(),
			Source:      constant.KafkaSourceAuthService,
		},
		payload: UserVerifiedPayload{
			UserID: userID,
			Email:  email,
		},
	}
}

// UserVerifiedProducer is responsible for producing product created events.
type UserVerifiedProducer struct {
	Producer  *mq.KafkaAsyncProducer
	RetryChan chan *sarama.ProducerMessage
	topic     string
}

// NewUserVerifiedProducer creates a new instance of UserVerifiedProducer.
func NewUserVerifiedProducer(producer *mq.KafkaAsyncProducer) mq.KafkaProducer {
	pr := &UserVerifiedProducer{
		Producer:  producer,
		topic:     constant.UserVerificationTopic,
		RetryChan: make(chan *sarama.ProducerMessage, 100),
	}

	return pr
}

// Send implements the KafkaProducer interface.
func (p *UserVerifiedProducer) Send(ctx context.Context, event mq.BaseEvent) error {
	return p.Producer.ProduceAsync(ctx, p.topic, event)
}

// Topic returns the topic name.
func (p *UserVerifiedProducer) Topic() string {
	return p.topic
}
