package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/entity"
)

// InboxRepository defines the methods for interacting with the inbox.
type InboxRepository interface {
	// Create inserts a new inbox event or returns existing if duplicate message_id.
	Create(ctx context.Context, event *entity.InboxEvent) (*entity.InboxEvent, error)
	// GetEventsForProcessing retrieves events that are ready for processing.
	GetEventsForProcessing(ctx context.Context, limit int64) ([]*entity.InboxEvent, error)
	// MarkAsProcessing updates an event status to processing.
	MarkAsProcessing(ctx context.Context, id uuid.UUID) error
	// MarkAsProcessed updates an event status to processed.
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
	// MarkAsFailed updates an event status to failed.
	MarkAsFailed(ctx context.Context, id uuid.UUID, errorMsg string) error
	// ScheduleForRetry schedules an event for retry.
	ScheduleForRetry(
		ctx context.Context,
		id uuid.UUID,
		errorMsg string,
		scheduledFor time.Time,
	) error
	// GetEventByID retrieves an event by its ID.
	GetEventByID(ctx context.Context, id uuid.UUID) (*entity.InboxEvent, error)
	// GetEventByMessageID retrieves an event by its message ID for duplicate detection.
	GetEventByMessageID(ctx context.Context, messageID uuid.UUID) (*entity.InboxEvent, error)
	// IncrementAttempts increments the attempt counter for an event.
	IncrementAttempts(ctx context.Context, id uuid.UUID) error
	// CleanupProcessedEvents removes processed events older than the specified duration.
	CleanupProcessedEvents(ctx context.Context, olderThan time.Duration) error
}

// inboxRepository implements the InboxRepository.
type inboxRepository struct {
	db DBTX
}

// NewInboxRepository creates a new instance of inboxRepository.
func NewInboxRepository(db DBTX) InboxRepository {
	return &inboxRepository{
		db: db,
	}
}

// Create inserts a new inbox event or returns existing if duplicate message_id.
func (r *inboxRepository) Create(
	ctx context.Context,
	event *entity.InboxEvent,
) (*entity.InboxEvent, error) {
	// First try to get existing event by message_id to handle duplicates
	existingEvent, err := r.GetEventByMessageID(ctx, event.MessageID)
	if err != nil && err.Error() != constant.InboxEventNotFoundErrorMessage {
		return nil, fmt.Errorf("failed to check for existing event: %w", err)
	}

	if existingEvent != nil {
		// Mark as duplicate and return existing event
		existingEvent.MarkAsDuplicate()

		if err = r.updateEventStatus(ctx, existingEvent.ID, constant.InboxStatusDuplicate, nil); err != nil {
			return nil, fmt.Errorf("failed to mark event as duplicate: %w", err)
		}

		return existingEvent, nil
	}

	// Insert new event
	query := `
		INSERT INTO inbox_events (
			id, message_id, aggregate_type, aggregate_id, event_type, topic, 
			source_service, payload, status, created_at, scheduled_for, attempts,
			correlation_id, causation_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, message_id, aggregate_type, aggregate_id, event_type, topic,
			source_service, payload, status, created_at, processed_at, scheduled_for,
			attempts, last_error, correlation_id, causation_id
	`

	row := r.db.QueryRow(
		ctx,
		query,
		event.ID,
		event.MessageID,
		event.AggregateType,
		event.AggregateID,
		event.EventType,
		event.Topic,
		event.SourceService,
		event.Payload,
		string(event.Status),
		event.CreatedAt,
		event.ScheduledFor,
		event.Attempts,
		event.CorrelationID,
		event.CausationID,
	)

	createdEvent, err := scanInboxEvent(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create inbox event: %w", err)
	}

	return createdEvent, nil
}

// GetEventsForProcessing retrieves events ready for processing.
func (r *inboxRepository) GetEventsForProcessing(
	ctx context.Context,
	limit int64,
) ([]*entity.InboxEvent, error) {
	query := `
		SELECT 
			id, message_id, aggregate_type, aggregate_id, event_type, topic,
			source_service, payload, status, created_at, processed_at, scheduled_for,
			attempts, last_error, correlation_id, causation_id
		FROM inbox_events 
		WHERE status IN ('pending', 'retry') 
		AND scheduled_for <= $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query inbox events: %w", err)
	}
	defer rows.Close()

	var events []*entity.InboxEvent

	for rows.Next() {
		event, errEvt := scanInboxEvent(rows)
		if errEvt != nil {
			return nil, fmt.Errorf("failed to scan inbox event: %w", errEvt)
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inbox events: %w", err)
	}

	return events, nil
}

// MarkAsProcessing updates an event status to processing.
func (r *inboxRepository) MarkAsProcessing(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE inbox_events 
		SET status = 'processing', attempts = attempts + 1
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark event as processing: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found: %s", id)
	}

	return nil
}

// MarkAsProcessed updates an event status to processed.
func (r *inboxRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	return r.updateEventStatus(ctx, id, constant.InboxStatusProcessed, nil)
}

// MarkAsFailed updates an event status to failed.
func (r *inboxRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errorMsg string) error {
	return r.updateEventStatus(ctx, id, constant.InboxStatusFailed, &errorMsg)
}

// ScheduleForRetry schedules an event for retry.
func (r *inboxRepository) ScheduleForRetry(
	ctx context.Context,
	id uuid.UUID,
	errorMsg string,
	scheduledFor time.Time,
) error {
	query := `
		UPDATE inbox_events 
		SET status = 'retry', scheduled_for = $1, last_error = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(ctx, query, scheduledFor, errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to schedule event for retry: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found: %s", id)
	}

	return nil
}

// GetEventByID retrieves an event by its ID.
func (r *inboxRepository) GetEventByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.InboxEvent, error) {
	query := `
		SELECT 
			id, message_id, aggregate_type, aggregate_id, event_type, topic,
			source_service, payload, status, created_at, processed_at, scheduled_for,
			attempts, last_error, correlation_id, causation_id
		FROM inbox_events 
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	return scanInboxEvent(row)
}

// GetEventByMessageID retrieves an event by its message ID for duplicate detection.
func (r *inboxRepository) GetEventByMessageID(
	ctx context.Context,
	messageID uuid.UUID,
) (*entity.InboxEvent, error) {
	query := `
		SELECT 
			id, message_id, aggregate_type, aggregate_id, event_type, topic,
			source_service, payload, status, created_at, processed_at, scheduled_for,
			attempts, last_error, correlation_id, causation_id
		FROM inbox_events 
		WHERE message_id = $1
	`

	row := r.db.QueryRow(ctx, query, messageID)

	event, err := scanInboxEvent(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.InboxEventNotFoundErrorMessage)
		}

		return nil, err
	}

	return event, nil
}

// IncrementAttempts increments the attempt count for an event.
func (r *inboxRepository) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE inbox_events 
		SET attempts = attempts + 1
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found: %s", id)
	}

	return nil
}

// CleanupProcessedEvents removes processed events older than specified duration.
func (r *inboxRepository) CleanupProcessedEvents(
	ctx context.Context,
	olderThan time.Duration,
) error {
	cutoffTime := time.Now().Add(-olderThan)

	query := `
		DELETE FROM inbox_events 
		WHERE status IN ('processed', 'duplicate') 
		AND processed_at < $1
	`

	_, err := r.db.Exec(ctx, query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to cleanup processed events: %w", err)
	}

	return nil
}

// updateEventStatus is a helper function to update event status.
func (r *inboxRepository) updateEventStatus(
	ctx context.Context,
	id uuid.UUID,
	status constant.InboxStatus,
	errorMsg *string,
) error {
	query := `
		UPDATE inbox_events 
		SET status = $2, 
			processed_at = CASE WHEN $2 = 'processed' THEN NOW() ELSE processed_at END,
			last_error = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, string(status), errorMsg)
	if err != nil {
		return fmt.Errorf("failed to update event status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found: %s", id)
	}

	return nil
}

// scanInboxEvent scans a database row into an InboxEvent struct.
func scanInboxEvent(row pgx.Row) (*entity.InboxEvent, error) {
	var event entity.InboxEvent

	var statusStr string

	// Use nullable types for columns that can be NULL
	var processedAt, scheduledFor pgtype.Timestamptz

	var lastError pgtype.Text

	var correlationID, causationID pgtype.UUID

	err := row.Scan(
		&event.ID,
		&event.MessageID,
		&event.AggregateType,
		&event.AggregateID,
		&event.EventType,
		&event.Topic,
		&event.SourceService,
		&event.Payload,
		&statusStr,
		&event.CreatedAt,
		&processedAt,
		&scheduledFor,
		&event.Attempts,
		&lastError,
		&correlationID,
		&causationID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.InboxEventNotFoundErrorMessage)
		}

		return nil, err
	}

	event.Status = constant.InboxStatus(statusStr)

	// Handle nullable processed_at
	if processedAt.Status == pgtype.Present {
		event.ProcessedAt = &processedAt.Time
	} else {
		event.ProcessedAt = nil
	}

	// Handle nullable scheduled_for (though it shouldn't be null)
	if scheduledFor.Status == pgtype.Present {
		event.ScheduledFor = scheduledFor.Time
	} else {
		// Fallback to current time if null
		event.ScheduledFor = time.Now()
	}

	// Handle nullable last_error
	if lastError.Status == pgtype.Present {
		event.LastError = &lastError.String
	} else {
		event.LastError = nil
	}

	// Handle nullable correlation_id
	if correlationID.Status == pgtype.Present {
		uuidValue := uuid.UUID(correlationID.Bytes)
		event.CorrelationID = &uuidValue
	} else {
		event.CorrelationID = nil
	}

	// Handle nullable causation_id
	if causationID.Status == pgtype.Present {
		uuidValue := uuid.UUID(causationID.Bytes)
		event.CausationID = &uuidValue
	} else {
		event.CausationID = nil
	}

	return &event, nil
}
