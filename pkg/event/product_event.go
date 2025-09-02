// Package event provides the commont event definitions for microservices.
package event

// ProductCreatedEvent is the envelope for product creation events.
type ProductCreatedEvent struct {
	ProductID string `json:"product_id"`
}
