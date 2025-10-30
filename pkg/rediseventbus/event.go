package rediseventbus

import (
	"fmt"
	"time"

	"github.com/bytedance/sonic"
)

// BaseEvent provides common fields for all events.
type BaseEvent struct {
	SourceInstanceID string    `json:"source_instance_id"`
	EventType        string    `json:"event_type"`
	Timestamp        time.Time `json:"timestamp"`
	Payload          []byte    `json:"payload"`
}

// NewBaseEvent creates a new base event.
func NewBaseEvent(sourceInstanceID, eventType string, payload interface{}) (*BaseEvent, error) {
	payloadBytes, err := sonic.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event payload: %w", err)
	}

	return &BaseEvent{
		SourceInstanceID: sourceInstanceID,
		EventType:        eventType,
		Timestamp:        time.Now(),
		Payload:          payloadBytes,
	}, nil
}

// GetSourceInstanceID returns the source instance ID.
func (e *BaseEvent) GetSourceInstanceID() string {
	return e.SourceInstanceID
}

// GetType returns the event type.
func (e *BaseEvent) GetType() string {
	return e.EventType
}

// Marshal serializes the event to JSON bytes.
func (e *BaseEvent) Marshal() ([]byte, error) {
	data, err := sonic.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	return data, nil
}

// Unmarshal deserializes an event from JSON bytes.
func Unmarshal(data []byte) (*BaseEvent, error) {
	var event BaseEvent
	if err := sonic.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	return &event, nil
}

// UnmarshalPayload deserializes the event payload into the target.
func (e *BaseEvent) UnmarshalPayload(target interface{}) error {
	if err := sonic.Unmarshal(e.Payload, target); err != nil {
		return fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	return nil
}
