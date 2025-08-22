// Package event provides the event definitions and handlers for the order service.
package event

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/shopspring/decimal"
)

// ProductCreatedPayload holds the data for the product created event.
type ProductCreatedPayload struct {
	ProductID uuid.UUID       `json:"product_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Quantity  int             `json:"quantity"`
}

// ProductCreatedEvent is the envelope for all product events.
type ProductCreatedEvent struct {
	Metadata mq.KafkaMetadata      `json:"metadata"`
	Payload  ProductCreatedPayload `json:"payload"`
}
