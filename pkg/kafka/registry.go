package kafka

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// EventRegistry maps event types to their concrete implementations.
type EventRegistry struct {
	eventTypes map[string]reflect.Type
}

// NewEventRegistry creates a new event registry.
func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		eventTypes: make(map[string]reflect.Type),
	}
}

// Register registers an event type with the registry.
func (r *EventRegistry) Register(eventType string, event BaseEvent) {
	r.eventTypes[eventType] = reflect.TypeOf(event).Elem()
}

// CreateEvent creates a new event instance by type.
func (r *EventRegistry) CreateEvent(eventType string) (BaseEvent, error) {
	eventTypeReflect, exists := r.eventTypes[eventType]
	if !exists {
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}

	// Create new instance
	eventValue := reflect.New(eventTypeReflect)

	event, ok := eventValue.Interface().(BaseEvent)
	if !ok {
		return nil, fmt.Errorf("event type %s does not implement BaseEvent", eventType)
	}

	return event, nil
}

// UnmarshalEvent unmarshals JSON payload into the correct event type.
func (r *EventRegistry) UnmarshalEvent(eventType string, payload []byte) (BaseEvent, error) {
	event, err := r.CreateEvent(eventType)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(payload, event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	return event, nil
}
