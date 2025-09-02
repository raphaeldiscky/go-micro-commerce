// Package event defines domain events for the product service.
package event

import "github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

type (
	// BaseEvent defines the interface for all events in the product service.
	BaseEvent = kafka.BaseEvent
	// KafkaMetadata provides common event properties.
	KafkaMetadata = kafka.Metadata
)

// Producer defines the interface for producing events.
type Producer interface {
	Produce(topic string, event BaseEvent) error
}
