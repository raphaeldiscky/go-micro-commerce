package events

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

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	Publish(event DomainEvent) error
}

// SellerCreatedEvent represents a seller creation event
type SellerCreatedEvent struct {
	BaseEvent
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

func (e SellerCreatedEvent) GetData() interface{} {
	return map[string]interface{}{
		"name":    e.Name,
		"email":   e.Email,
		"phone":   e.Phone,
		"address": e.Address,
	}
}

// NewSellerCreatedEvent creates a new seller created event
func NewSellerCreatedEvent(sellerId uuid.UUID, name, email, phone, address string) *SellerCreatedEvent {
	return &SellerCreatedEvent{
		BaseEvent: BaseEvent{
			EventId:     uuid.New(),
			EventType:   "SellerCreated",
			AggregateId: sellerId,
			OccurredAt:  time.Now(),
		},
		Name:    name,
		Email:   email,
		Phone:   phone,
		Address: address,
	}
}

// SellerUpdatedEvent represents a seller update event
type SellerUpdatedEvent struct {
	BaseEvent
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

func (e SellerUpdatedEvent) GetData() interface{} {
	return map[string]interface{}{
		"name":    e.Name,
		"email":   e.Email,
		"phone":   e.Phone,
		"address": e.Address,
	}
}

// NewSellerUpdatedEvent creates a new seller updated event
func NewSellerUpdatedEvent(sellerId uuid.UUID, name, email, phone, address string) *SellerUpdatedEvent {
	return &SellerUpdatedEvent{
		BaseEvent: BaseEvent{
			EventId:     uuid.New(),
			EventType:   "SellerUpdated",
			AggregateId: sellerId,
			OccurredAt:  time.Now(),
		},
		Name:    name,
		Email:   email,
		Phone:   phone,
		Address: address,
	}
}

// SellerDeletedEvent represents a seller deletion event
type SellerDeletedEvent struct {
	BaseEvent
}

func (e SellerDeletedEvent) GetData() interface{} {
	return map[string]interface{}{
		"deleted_at": e.OccurredAt,
	}
}

// NewSellerDeletedEvent creates a new seller deleted event
func NewSellerDeletedEvent(sellerId uuid.UUID) *SellerDeletedEvent {
	return &SellerDeletedEvent{
		BaseEvent: BaseEvent{
			EventId:     uuid.New(),
			EventType:   "SellerDeleted",
			AggregateId: sellerId,
			OccurredAt:  time.Now(),
		},
	}
}

// SellerActivatedEvent represents a seller activation event
type SellerActivatedEvent struct {
	BaseEvent
}

func (e SellerActivatedEvent) GetData() interface{} {
	return map[string]interface{}{
		"activated_at": e.OccurredAt,
	}
}

// NewSellerActivatedEvent creates a new seller activated event
func NewSellerActivatedEvent(sellerId uuid.UUID) *SellerActivatedEvent {
	return &SellerActivatedEvent{
		BaseEvent: BaseEvent{
			EventId:     uuid.New(),
			EventType:   "SellerActivated",
			AggregateId: sellerId,
			OccurredAt:  time.Now(),
		},
	}
}

// SellerDeactivatedEvent represents a seller deactivation event
type SellerDeactivatedEvent struct {
	BaseEvent
}

func (e SellerDeactivatedEvent) GetData() interface{} {
	return map[string]interface{}{
		"deactivated_at": e.OccurredAt,
	}
}

// NewSellerDeactivatedEvent creates a new seller deactivated event
func NewSellerDeactivatedEvent(sellerId uuid.UUID) *SellerDeactivatedEvent {
	return &SellerDeactivatedEvent{
		BaseEvent: BaseEvent{
			EventId:     uuid.New(),
			EventType:   "SellerDeactivated",
			AggregateId: sellerId,
			OccurredAt:  time.Now(),
		},
	}
}
