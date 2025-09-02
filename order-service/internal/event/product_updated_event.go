// Package event provides the event definitions and handlers for the order service.
package event

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/shopspring/decimal"
)

// ProductUpdatedPayload holds the data for the product updated event.
type ProductUpdatedPayload struct {
	ProductID uuid.UUID       `json:"product_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Quantity  int64           `json:"quantity"`
}

// ProductUpdatedEvent is the envelope for all product events.
type ProductUpdatedEvent struct {
	Metadata kafka.Metadata        `json:"metadata"`
	Payload  ProductUpdatedPayload `json:"payload"`
}
