// Package event provides the event definitions and handlers for the order service.
package event

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
)

// ProductDeletedPayload holds the data for the product deleted event.
type ProductDeletedPayload struct {
	ProductID uuid.UUID `json:"product_id"`
}

// ProductDeletedEvent is the envelope for all product events.
type ProductDeletedEvent struct {
	Metadata mq.KafkaMetadata      `json:"metadata"`
	Payload  ProductDeletedPayload `json:"payload"`
}
