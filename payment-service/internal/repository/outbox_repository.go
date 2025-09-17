package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
)

// OutboxRepository defines the methods for interacting with the outbox.
type OutboxRepository interface {
	// Create inserts a new outbox event.
	Create(ctx context.Context, event *entity.OutboxEvent) error
	// GetEventsForProcessing retrieves events that are ready for processing.
	GetEventsForProcessing(ctx context.Context, limit int64) ([]*entity.OutboxEvent, error)
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
	GetEventByID(ctx context.Context, id uuid.UUID) (*entity.OutboxEvent, error)
	// IncrementAttempts increments the attempt counter for an event.
	IncrementAttempts(ctx context.Context, id uuid.UUID) error
	// CleanupProcessedEvents removes processed events older than the specified duration.
	CleanupProcessedEvents(ctx context.Context, olderThan time.Duration) error
}

// outboxRepository implements the OutboxRepository.
type outboxRepository struct {
	db DBTX
}

// NewOutboxRepository creates a new instance of OutboxRepositoryPostgres.
func NewOutboxRepository(db DBTX) OutboxRepository {
	return &outboxRepository{
		db: db,
	}
}

// Create inserts a new outbox event.
func (r *outboxRepository) Create(ctx context.Context, event *entity.OutboxEvent) error {
	query := `
		INSERT INTO outbox_events (
			id, aggregate_type, aggregate_id, event_type, topic, 
			payload, status, created_at, scheduled_for, attempts
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		event.ID,
		event.AggregateType,
		event.AggregateID,
		event.EventType,
		event.Topic,
		event.Payload,
		string(event.Status),
		event.CreatedAt,
		event.ScheduledFor,
		event.Attempts,
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox event: %w", err)
	}

	return nil
}

// GetEventsForProcessing retrieves events ready for processing.
func (r *outboxRepository) GetEventsForProcessing(
	ctx context.Context,
	limit int64,
) ([]*entity.OutboxEvent, error) {
	query := `
		SELECT 
			id, aggregate_type, aggregate_id, event_type, topic,
			payload, status, created_at, processed_at, scheduled_for,
			attempts, last_error
		FROM outbox_events 
		WHERE status IN ('pending', 'retry') 
		AND scheduled_for <= $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbox events: %w", err)
	}
	defer rows.Close()

	var events []*entity.OutboxEvent

	for rows.Next() {
		event, errEvt := scanOutboxEvent(rows)
		if errEvt != nil {
			return nil, fmt.Errorf("failed to scan outbox event: %w", errEvt)
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating outbox events: %w", err)
	}

	return events, nil
}

// MarkAsProcessing updates an event status to processing.
func (r *outboxRepository) MarkAsProcessing(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE outbox_events 
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
func (r *outboxRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE outbox_events 
		SET status = 'processed', processed_at = $1, last_error = NULL
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found: %s", id)
	}

	return nil
}

// MarkAsFailed updates an event status to failed.
func (r *outboxRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errorMsg string) error {
	query := `
		UPDATE outbox_events 
		SET status = 'failed', last_error = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to mark event as failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found: %s", id)
	}

	return nil
}

// ScheduleForRetry schedules an event for retry.
func (r *outboxRepository) ScheduleForRetry(
	ctx context.Context,
	id uuid.UUID,
	errorMsg string,
	scheduledFor time.Time,
) error {
	query := `
		UPDATE outbox_events 
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
func (r *outboxRepository) GetEventByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.OutboxEvent, error) {
	query := `
		SELECT 
			id, aggregate_type, aggregate_id, event_type, topic,
			payload, status, created_at, processed_at, scheduled_for,
			attempts, last_error
		FROM outbox_events 
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	return scanOutboxEvent(row)
}

// IncrementAttempts increments the attempt count for an event.
func (r *outboxRepository) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE outbox_events 
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
func (r *outboxRepository) CleanupProcessedEvents(
	ctx context.Context,
	olderThan time.Duration,
) error {
	cutoffTime := time.Now().Add(-olderThan)

	query := `
		DELETE FROM outbox_events 
		WHERE status = 'processed' 
		AND processed_at < $1
	`

	_, err := r.db.Exec(ctx, query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to cleanup processed events: %w", err)
	}

	return nil
}

// scanOutboxEvent scans a database row into an OutboxEvent struct.
func scanOutboxEvent(row pgx.Row) (*entity.OutboxEvent, error) {
	var event entity.OutboxEvent

	var statusStr string

	// Use nullable types for columns that can be NULL
	var processedAt, scheduledFor pgtype.Timestamptz

	var lastError pgtype.Text

	err := row.Scan(
		&event.ID,
		&event.AggregateType,
		&event.AggregateID,
		&event.EventType,
		&event.Topic,
		&event.Payload,
		&statusStr,
		&event.CreatedAt,
		&processedAt,  // Now using pgtype.Timestamptz
		&scheduledFor, // Now using pgtype.Timestamptz
		&event.Attempts,
		&lastError,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.OutboxEventNotFoundErrorMessage)
		}

		return nil, err
	}

	event.Status = constant.OutboxStatus(statusStr)

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

	return &event, nil
}
