package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// OutboxEvent represents an event that is to be processed by the outbox.
type OutboxEvent struct {
	ID            uuid.UUID
	AggregateType string
	AggregateID   uuid.UUID
	EventType     string
	Topic         string
	Payload       []byte
	Status        constant.OutboxStatus
	CreatedAt     time.Time
	ProcessedAt   *time.Time
	ScheduledFor  time.Time
	Attempts      int64
	LastError     *string
}
