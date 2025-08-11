// Package event defines the events for the notification service.
package event

import "github.com/raphaeldiscky/go-micro-template/pkg/mq"

// UserVerifiedEvent represents the user verified event payload.
type UserVerifiedEvent struct {
	Email string `json:"email"`
}

// UserVerifiedConsumer handles the logic for processing user verified events.
type UserVerifiedConsumer struct {
	Consumer *mq.KafkaConsumer
}
