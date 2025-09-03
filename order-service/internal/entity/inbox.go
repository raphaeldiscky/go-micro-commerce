// Package entity defines the InboxEvent entity and its validation logic.
package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// InboxEvent represents an event consumed from message brokers using the inbox pattern.
type InboxEvent struct {
	ID            uuid.UUID
	MessageID     uuid.UUID       // Unique identifier from Kafka message metadata
	AggregateType string          // Type of aggregate from source service (e.g., 'order', 'product')
	AggregateID   uuid.UUID       // ID of the aggregate from source service
	EventType     string          // Type of event from source service
	Topic         string          // Kafka topic from which event was consumed
	SourceService string          // Name of the microservice that published the event
	Payload       json.RawMessage // Complete event payload from source service
	Status        constant.InboxStatus
	CreatedAt     time.Time
	ProcessedAt   *time.Time
	ScheduledFor  time.Time
	Attempts      int64
	LastError     *string
	CorrelationID *uuid.UUID // For tracing requests across services
	CausationID   *uuid.UUID // For linking cause-and-effect events
}

// NewInboxEvent creates a new inbox event with validation.
func NewInboxEvent(
	messageID uuid.UUID,
	aggregateType string,
	aggregateID uuid.UUID,
	eventType string,
	topic string,
	sourceService string,
	payload json.RawMessage,
	correlationID *uuid.UUID,
	causationID *uuid.UUID,
) *InboxEvent {
	now := time.Now()

	return &InboxEvent{
		ID:            uuid.New(),
		MessageID:     messageID,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Topic:         topic,
		SourceService: sourceService,
		Payload:       payload,
		Status:        constant.InboxStatusPending,
		CreatedAt:     now,
		ProcessedAt:   nil,
		ScheduledFor:  now,
		Attempts:      0,
		LastError:     nil,
		CorrelationID: correlationID,
		CausationID:   causationID,
	}
}

// MarkAsProcessing updates the event status to processing.
func (e *InboxEvent) MarkAsProcessing() {
	e.Status = constant.InboxStatusProcessing
	e.Attempts++
}

// MarkAsProcessed updates the event status to processed.
func (e *InboxEvent) MarkAsProcessed() {
	now := time.Now()
	e.Status = constant.InboxStatusProcessed
	e.ProcessedAt = &now
	e.LastError = nil
}

// MarkAsFailed updates the event status to failed with error message.
func (e *InboxEvent) MarkAsFailed(errorMsg string) {
	e.Status = constant.InboxStatusFailed
	e.LastError = &errorMsg
}

// ScheduleForRetry schedules the event for retry with exponential backoff.
func (e *InboxEvent) ScheduleForRetry(errorMsg string, scheduledFor time.Time) {
	e.Status = constant.InboxStatusRetry
	e.LastError = &errorMsg
	e.ScheduledFor = scheduledFor
}

// MarkAsDuplicate marks the event as a duplicate.
func (e *InboxEvent) MarkAsDuplicate() {
	e.Status = constant.InboxStatusDuplicate
}

// IsReadyForProcessing checks if the event is ready to be processed.
func (e *InboxEvent) IsReadyForProcessing() bool {
	return (e.Status == constant.InboxStatusPending || e.Status == constant.InboxStatusRetry) &&
		e.ScheduledFor.Before(time.Now().Add(1*time.Second))
}

// CanBeRetried checks if the event can be retried based on attempts.
func (e *InboxEvent) CanBeRetried(maxAttempts int64) bool {
	return e.Attempts < maxAttempts
}
