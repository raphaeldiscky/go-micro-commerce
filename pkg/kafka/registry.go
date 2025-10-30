package kafka

import (
	"fmt"
	"reflect"

	"github.com/bytedance/sonic"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafkaevent"
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
func (r *EventRegistry) Register(eventType string, evt kafkaevent.BaseEvent) {
	r.eventTypes[eventType] = reflect.TypeOf(evt).Elem()
}

// CreateEvent creates a new event instance by type.
func (r *EventRegistry) CreateEvent(eventType string) (kafkaevent.BaseEvent, error) {
	eventTypeReflect, exists := r.eventTypes[eventType]
	if !exists {
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}

	// Create new instance
	eventValue := reflect.New(eventTypeReflect)

	evt, ok := eventValue.Interface().(kafkaevent.BaseEvent)
	if !ok {
		return nil, fmt.Errorf("event type %s does not implement BaseEvent", eventType)
	}

	return evt, nil
}

// UnmarshalEvent unmarshals JSON payload into the correct event type.
func (r *EventRegistry) UnmarshalEvent(
	eventType string,
	payload []byte,
) (kafkaevent.BaseEvent, error) {
	evt, err := r.CreateEvent(eventType)
	if err != nil {
		return nil, err
	}

	if err = sonic.Unmarshal(payload, evt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event payload: %w", err)
	}

	return evt, nil
}
