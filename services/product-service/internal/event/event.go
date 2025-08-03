// Package event defines domain events for the product service.
package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DomainEvent represents a domain event interface.
type DomainEvent interface {
	GetEventID() uuid.UUID
	GetEventType() string
	GetAggregateID() uuid.UUID
	GetOccurredAt() time.Time
	GetData() interface{}
}

// BaseEvent provides common event properties.
type BaseEvent struct {
	EventID     uuid.UUID `json:"event_id"`
	EventType   string    `json:"event_type"`
	AggregateID uuid.UUID `json:"aggregate_id"`
	OccurredAt  time.Time `json:"occurred_at"`
	Source      string    `json:"source,omitempty"` // Service that produced the event
}

// ProductCreatedData holds the data for the product created event.
type ProductCreatedData struct {
	ProductID uuid.UUID       `json:"product_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
}

// ProductUpdatedData holds the data for the product updated event.
type ProductUpdatedData struct {
	ProductID    uuid.UUID          `json:"product_id"`
	Name         string             `json:"name"`
	Price        decimal.Decimal    `json:"price"`
	PreviousData ProductCreatedData `json:"previous_data,omitempty"` // Optional field for previous product data
}

// ProductDeletedData represents when a product is deleted.
type ProductDeletedData struct {
	ProductID uuid.UUID `json:"product_id"`
}

// ProductCreatedEvent is the envelope for all product events.
type ProductCreatedEvent struct {
	BaseEvent
	Data ProductCreatedData
}

// GetData returns the data associated with the ProductCreatedEvent.
func (e *ProductCreatedEvent) GetData() interface{} {
	return e.Data
}

// ProductUpdatedEvent is the envelope for product update events.
type ProductUpdatedEvent struct {
	BaseEvent
	Data ProductUpdatedData
}

// GetData returns the data associated with the ProductUpdatedEvent.
func (e *ProductUpdatedEvent) GetData() interface{} {
	return e.Data
}

// ProductDeletedEvent is the envelope for product deletion events.
type ProductDeletedEvent struct {
	BaseEvent
	Data ProductDeletedData
}

// GetData returns the data associated with the ProductDeletedEvent.
func (e *ProductDeletedEvent) GetData() interface{} {
	return e.Data
}

// GetEventID returns the unique identifier of the event.
func (e *BaseEvent) GetEventID() uuid.UUID { return e.EventID }

// GetEventType returns the type of the event.
func (e *BaseEvent) GetEventType() string { return e.EventType }

// GetAggregateID returns the identifier of the aggregate that this event belongs to.
func (e *BaseEvent) GetAggregateID() uuid.UUID { return e.AggregateID }

// GetOccurredAt returns the timestamp when the event occurred.
func (e *BaseEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// Publisher defines the interface for publishing event.
type Publisher interface {
	Publish(event DomainEvent) error
}
