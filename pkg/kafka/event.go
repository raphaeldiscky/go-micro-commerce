package kafka

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// GenericEvent wraps any event with metadata.
type GenericEvent struct {
	Metadata Metadata        `json:"metadata"`
	Payload  json.RawMessage `json:"payload"`
}

// BaseEvent represents a base event interface.
type BaseEvent interface {
	GetMetadata() Metadata
	GetPayload() interface{}
}

// Metadata provides common event properties.
type Metadata struct {
	EventID     uuid.UUID `json:"event_id"`
	EventType   string    `json:"event_type"`
	AggregateID uuid.UUID `json:"aggregate_id"`
	OccurredAt  time.Time `json:"occurred_at"`
	Source      string    `json:"source,omitempty"` // Service that produced the event
}
