package event

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event interface
type DomainEvent interface {
	GetEventId() uuid.UUID
	GetEventType() string
	GetAggregateId() uuid.UUID
	GetOccurredAt() time.Time
	GetData() interface{}
}

// BaseEvent provides common event properties
type BaseEvent struct {
	EventId     uuid.UUID
	EventType   string
	AggregateId uuid.UUID
	OccurredAt  time.Time
}

func (e BaseEvent) GetEventId() uuid.UUID     { return e.EventId }
func (e BaseEvent) GetEventType() string      { return e.EventType }
func (e BaseEvent) GetAggregateId() uuid.UUID { return e.AggregateId }
func (e BaseEvent) GetOccurredAt() time.Time  { return e.OccurredAt }

// ProductCreatedEvent represents when a product is created
type ProductCreatedEvent struct {
	BaseEvent
	ProductId uuid.UUID
	Name      string
	Price     float64
	SellerId  uuid.UUID
}

func NewProductCreatedEvent(productId uuid.UUID, name string, price float64, sellerId uuid.UUID) *ProductCreatedEvent {
	return &ProductCreatedEvent{
		BaseEvent: BaseEvent{
			EventId:     uuid.New(),
			EventType:   "ProductCreated",
			AggregateId: productId,
			OccurredAt:  time.Now(),
		},
		ProductId: productId,
		Name:      name,
		Price:     price,
		SellerId:  sellerId,
	}
}

func (e ProductCreatedEvent) GetData() interface{} {
	return struct {
		ProductId uuid.UUID `json:"product_id"`
		Name      string    `json:"name"`
		Price     float64   `json:"price"`
		SellerId  uuid.UUID `json:"seller_id"`
	}{
		ProductId: e.ProductId,
		Name:      e.Name,
		Price:     e.Price,
		SellerId:  e.SellerId,
	}
}

// ProductUpdatedEvent represents when a product is updated
type ProductUpdatedEvent struct {
	BaseEvent
	ProductId uuid.UUID
	Name      string
	Price     float64
	SellerId  uuid.UUID
}

func NewProductUpdatedEvent(productId uuid.UUID, name string, price float64, sellerId uuid.UUID) *ProductUpdatedEvent {
	return &ProductUpdatedEvent{
		BaseEvent: BaseEvent{
			EventId:     uuid.New(),
			EventType:   "ProductUpdated",
			AggregateId: productId,
			OccurredAt:  time.Now(),
		},
		ProductId: productId,
		Name:      name,
		Price:     price,
		SellerId:  sellerId,
	}
}

func (e ProductUpdatedEvent) GetData() interface{} {
	return struct {
		ProductId uuid.UUID `json:"product_id"`
		Name      string    `json:"name"`
		Price     float64   `json:"price"`
		SellerId  uuid.UUID `json:"seller_id"`
	}{
		ProductId: e.ProductId,
		Name:      e.Name,
		Price:     e.Price,
		SellerId:  e.SellerId,
	}
}

// ProductDeletedEvent represents when a product is deleted
type ProductDeletedEvent struct {
	BaseEvent
	ProductId uuid.UUID
}

func NewProductDeletedEvent(productId uuid.UUID) *ProductDeletedEvent {
	return &ProductDeletedEvent{
		BaseEvent: BaseEvent{
			EventId:     uuid.New(),
			EventType:   "ProductDeleted",
			AggregateId: productId,
			OccurredAt:  time.Now(),
		},
		ProductId: productId,
	}
}

func (e ProductDeletedEvent) GetData() interface{} {
	return struct {
		ProductId uuid.UUID `json:"product_id"`
	}{
		ProductId: e.ProductId,
	}
}

// EventPublisher defines the interface for publishing event
type EventPublisher interface {
	Publish(event DomainEvent) error
}
