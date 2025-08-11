// Package event defines the events for the notification service.
package event

import "github.com/raphaeldiscky/go-micro-template/pkg/mq"

type UserVerifiedEvent struct {
	Email string `json:"email"`
}

type UserVerifiedConsumer struct {
	Consumer *mq.KafkaConsumer
}
