// Package event defines domain events for the product service.
package event

import (
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
)

type (
	// BaseEvent defines the interface for all events in the product service.
	BaseEvent = mq.BaseEvent
	// KafkaMetadata provides common event properties.
	KafkaMetadata = mq.KafkaMetadata
)

// AsyncProducer defines the interface for producing events.
type AsyncProducer interface {
	ProduceAsync(topic string, event BaseEvent) error
}
