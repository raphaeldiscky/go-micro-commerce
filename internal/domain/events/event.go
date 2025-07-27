package events

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event in the system.
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	AggregateType() string
	EventVersion() int
	OccurredAt() time.Time
	EventData() interface{}
}

// BaseDomainEvent provides a base implementation for domain events.
type BaseDomainEvent struct {
	ID                string      `json:"id"`
	EventTypeName     string      `json:"type"`
	AggregateId       string      `json:"aggregate_id"`
	AggregateTypeName string      `json:"aggregate_type"`
	Version           int         `json:"version"`
	Timestamp         time.Time   `json:"timestamp"`
	Data              interface{} `json:"data"`
}

// EventID returns the unique identifier of the event.
func (e BaseDomainEvent) EventID() string {
	return e.ID
}

// EventType returns the type of the event.
func (e BaseDomainEvent) EventType() string {
	return e.EventTypeName
}

// AggregateID returns the ID of the aggregate that this event belongs to.
func (e BaseDomainEvent) AggregateID() string {
	return e.AggregateId
}

// AggregateType returns the type of the aggregate that this event belongs to.
func (e BaseDomainEvent) AggregateType() string {
	return e.AggregateTypeName
}

// EventVersion returns the version of the event.
func (e BaseDomainEvent) EventVersion() int {
	return e.Version
}

// OccurredAt returns the time when the event occurred.
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// EventData returns the data associated with the event.
func (e BaseDomainEvent) EventData() interface{} {
	return e.Data
}

// NewBaseDomainEvent creates a new base domain event.
func NewBaseDomainEvent(
	eventType, aggregateID, aggregateType string,
	version int,
	data interface{},
) BaseDomainEvent {
	return BaseDomainEvent{
		ID:                uuid.New().String(),
		EventTypeName:     eventType,
		AggregateId:       aggregateID,
		AggregateTypeName: aggregateType,
		Version:           version,
		Timestamp:         time.Now().UTC(),
		Data:              data,
	}
}
