package events

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event in the system
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	AggregateType() string
	EventVersion() int
	OccurredAt() time.Time
	EventData() interface{}
}

// BaseDomainEvent provides a base implementation for domain events
type BaseDomainEvent struct {
	ID                string      `json:"id"`
	EventTypeName     string      `json:"type"`
	AggregateId       string      `json:"aggregate_id"`
	AggregateTypeName string      `json:"aggregate_type"`
	Version           int         `json:"version"`
	Timestamp         time.Time   `json:"timestamp"`
	Data              interface{} `json:"data"`
}

func (e BaseDomainEvent) EventID() string {
	return e.ID
}

func (e BaseDomainEvent) EventType() string {
	return e.EventTypeName
}

func (e BaseDomainEvent) AggregateID() string {
	return e.AggregateId
}

func (e BaseDomainEvent) AggregateType() string {
	return e.AggregateTypeName
}

func (e BaseDomainEvent) EventVersion() int {
	return e.Version
}

func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.Timestamp
}

func (e BaseDomainEvent) EventData() interface{} {
	return e.Data
}

// NewBaseDomainEvent creates a new base domain event
func NewBaseDomainEvent(eventType, aggregateID, aggregateType string, version int, data interface{}) BaseDomainEvent {
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
