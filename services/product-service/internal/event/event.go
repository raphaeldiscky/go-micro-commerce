// Package event defines domain events for the product service.
package event

import (
	"time"

	"github.com/google/uuid"
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
	EventID     uuid.UUID
	EventType   string
	AggregateID uuid.UUID
	OccurredAt  time.Time
}

// GetEventID returns the unique identifier of the event.
func (e BaseEvent) GetEventID() uuid.UUID { return e.EventID }

// GetEventType returns the type of the event.
func (e BaseEvent) GetEventType() string { return e.EventType }

// GetAggregateID returns the identifier of the aggregate that this event belongs to.
func (e BaseEvent) GetAggregateID() uuid.UUID { return e.AggregateID }

// GetOccurredAt returns the timestamp when the event occurred.
func (e BaseEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// ProductCreatedEvent represents when a product is created.
type ProductCreatedEvent struct {
	BaseEvent
	ProductID uuid.UUID
	Name      string
	Price     float64
}

// NewProductCreatedEvent creates a new ProductCreatedEvent.
func NewProductCreatedEvent(productID uuid.UUID, name string, price float64) *ProductCreatedEvent {
	return &ProductCreatedEvent{
		BaseEvent: BaseEvent{
			EventID:     uuid.New(),
			EventType:   "ProductCreated",
			AggregateID: productID,
			OccurredAt:  time.Now(),
		},
		ProductID: productID,
		Name:      name,
		Price:     price,
	}
}

// GetData returns the data associated with the ProductCreatedEvent.
func (e *ProductCreatedEvent) GetData() interface{} {
	return struct {
		ProductID uuid.UUID `json:"product_id"`
		Name      string    `json:"name"`
		Price     float64   `json:"price"`
	}{
		ProductID: e.ProductID,
		Name:      e.Name,
		Price:     e.Price,
	}
}

// ProductUpdatedEvent represents when a product is updated.
type ProductUpdatedEvent struct {
	BaseEvent
	ProductID uuid.UUID
	Name      string
	Price     float64
}

// NewProductUpdatedEvent creates a new ProductUpdatedEvent.
func NewProductUpdatedEvent(productID uuid.UUID, name string, price float64) *ProductUpdatedEvent {
	return &ProductUpdatedEvent{
		BaseEvent: BaseEvent{
			EventID:     uuid.New(),
			EventType:   "ProductUpdated",
			AggregateID: productID,
			OccurredAt:  time.Now(),
		},
		ProductID: productID,
		Name:      name,
		Price:     price,
	}
}

// GetData returns the data associated with the ProductUpdatedEvent.
func (e *ProductUpdatedEvent) GetData() interface{} {
	return struct {
		ProductID uuid.UUID `json:"product_id"`
		Name      string    `json:"name"`
		Price     float64   `json:"price"`
	}{
		ProductID: e.ProductID,
		Name:      e.Name,
		Price:     e.Price,
	}
}

// ProductDeletedEvent represents when a product is deleted.
type ProductDeletedEvent struct {
	BaseEvent
	ProductID uuid.UUID
}

// NewProductDeletedEvent creates a new ProductDeletedEvent.
func NewProductDeletedEvent(productID uuid.UUID) *ProductDeletedEvent {
	return &ProductDeletedEvent{
		BaseEvent: BaseEvent{
			EventID:     uuid.New(),
			EventType:   "ProductDeleted",
			AggregateID: productID,
			OccurredAt:  time.Now(),
		},
		ProductID: productID,
	}
}

// GetData returns the data associated with the ProductDeletedEvent.
func (e *ProductDeletedEvent) GetData() interface{} {
	return struct {
		ProductID uuid.UUID `json:"product_id"`
	}{
		ProductID: e.ProductID,
	}
}

// Publisher defines the interface for publishing event.
type Publisher interface {
	Publish(event DomainEvent) error
}
