// Package event defines domain events for the product service.
package event

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/shopspring/decimal"
)

type (
	// BaseEvent defines the interface for all events in the product service.
	BaseEvent = mq.BaseEvent
	// KafkaMetadata provides common event properties.
	KafkaMetadata = mq.KafkaMetadata
)

// ProductCreatedPayload holds the data for the product created event.
type ProductCreatedPayload struct {
	ProductID uuid.UUID       `json:"product_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Quantity  int             `json:"quantity"`
}

// ProductUpdatedPayload holds the data for the product updated event.
type ProductUpdatedPayload struct {
	ProductID    uuid.UUID             `json:"product_id"`
	Name         string                `json:"name"`
	Price        decimal.Decimal       `json:"price"`
	Quantity     int                   `json:"quantity"`
	PreviousData ProductCreatedPayload `json:"previous_data,omitempty"` // Optional field for previous product data
}

// ProductDeletedPayload represents when a product is deleted.
type ProductDeletedPayload struct {
	ProductID uuid.UUID `json:"product_id"`
}

// ProductCreatedEvent is the envelope for all product events.
type ProductCreatedEvent struct {
	Metadata KafkaMetadata
	Payload  ProductCreatedPayload
}

// GetPayload returns the data associated with the ProductCreatedEvent.
func (e *ProductCreatedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the ProductCreatedEvent.
func (e *ProductCreatedEvent) GetMetadata() KafkaMetadata {
	return e.Metadata
}

// ProductUpdatedEvent is the envelope for product update events.
type ProductUpdatedEvent struct {
	Metadata KafkaMetadata
	Payload  ProductUpdatedPayload
}

// GetPayload returns the data associated with the ProductUpdatedEvent.
func (e *ProductUpdatedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the ProductUpdatedEvent.
func (e *ProductUpdatedEvent) GetMetadata() KafkaMetadata {
	return e.Metadata
}

// ProductDeletedEvent is the envelope for product deletion events.
type ProductDeletedEvent struct {
	Metadata KafkaMetadata
	Payload  ProductDeletedPayload
}

// GetPayload returns the data associated with the ProductDeletedEvent.
func (e *ProductDeletedEvent) GetPayload() interface{} {
	return e.Payload
}

// GetMetadata returns the metadata associated with the ProductDeletedEvent.
func (e *ProductDeletedEvent) GetMetadata() KafkaMetadata {
	return e.Metadata
}

// Producer defines the interface for producing events.
type Producer interface {
	Produce(topic string, event BaseEvent) error
}
