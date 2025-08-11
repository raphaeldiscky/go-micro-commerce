package mq

import (
	"time"

	"github.com/google/uuid"
)

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
