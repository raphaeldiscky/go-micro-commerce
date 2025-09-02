// Package event provides the event definitions and handlers for the order service.
package event

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
)

// ProductDeletedPayload holds the data for the product deleted event.
type ProductDeletedPayload struct {
	ProductID uuid.UUID `json:"product_id"`
}

// ProductDeletedEvent is the envelope for all product events.
type ProductDeletedEvent struct {
	Metadata kafka.Metadata        `json:"metadata"`
	Payload  ProductDeletedPayload `json:"payload"`
}
