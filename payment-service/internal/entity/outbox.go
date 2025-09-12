package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
)

// OutboxEvent represents an event that is to be processed by the outbox.
type OutboxEvent struct {
	CreatedAt     time.Time
	ScheduledFor  time.Time
	ProcessedAt   *time.Time
	LastError     *string
	AggregateType string
	EventType     string
	Topic         string
	Status        constant.OutboxStatus
	Payload       []byte
	Attempts      int64
	ID            uuid.UUID
	AggregateID   uuid.UUID
}
