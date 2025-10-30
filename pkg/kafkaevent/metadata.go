package kafkaevent

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
)

// Metadata provides common event properties.
type Metadata struct {
	EventID     uuid.UUID `json:"event_id"`
	EventType   string    `json:"event_type"`
	AggregateID uuid.UUID `json:"aggregate_id"`
	OccurredAt  time.Time `json:"occurred_at"`
	Source      string    `json:"source,omitempty"` // Service that produced the event
}

// GenericEvent wraps any event with metadata.
type GenericEvent struct {
	Metadata Metadata               `json:"metadata"`
	Payload  sonic.NoCopyRawMessage `json:"payload"`
}

// BaseEvent represents a base event interface.
type BaseEvent interface {
	GetMetadata() Metadata
	GetPayload() any
}
