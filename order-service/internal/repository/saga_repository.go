package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// SagaStateRepositoryInterface interface for persisting saga state.
type SagaStateRepositoryInterface interface {
	// Create inserts a new saga state.
	Create(ctx context.Context, state *entity.SagaState) error
	// Update updates an existing saga state.
	Update(ctx context.Context, state *entity.SagaState) error
	// FindByID finds a saga state by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.SagaState, error)
	// FindByOrderID finds a saga state by order ID.
	FindByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.SagaState, error)
	// FindPendingOrFailed finds pending or failed sagas for recovery.
	FindPendingOrFailed(ctx context.Context, limit int64) ([]*entity.SagaState, error)
	// MarkAsExecuting updates saga status to executing.
	MarkAsExecuting(ctx context.Context, id uuid.UUID) error
	// MarkAsCompensating updates saga status to compensating.
	MarkAsCompensating(ctx context.Context, id uuid.UUID) error
	// MarkAsCompleted updates saga status to completed.
	MarkAsCompleted(ctx context.Context, id uuid.UUID) error
	// MarkAsFailed updates saga status to failed with error message.
	MarkAsFailed(ctx context.Context, id uuid.UUID, errorMsg string) error
	// MarkAsCompensated updates saga status to compensated.
	MarkAsCompensated(ctx context.Context, id uuid.UUID) error
}

// SagaStateRepository implements the SagaStateRepositoryInterface.
type SagaStateRepository struct {
	db DBTX
}

// NewSagaStateRepository creates a new instance of SagaStateRepository.
func NewSagaStateRepository(db DBTX) SagaStateRepositoryInterface {
	return &SagaStateRepository{
		db: db,
	}
}

// Create inserts a new saga state.
func (r *SagaStateRepository) Create(ctx context.Context, state *entity.SagaState) error {
	query := `
		INSERT INTO saga_states (
			id, order_id, status, current_step, 
			executed_steps, compensated_steps, data, error,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`

	executedStepsJSON, err := json.Marshal(state.ExecutedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal executed steps: %w", err)
	}

	compensatedStepsJSON, err := json.Marshal(state.CompensatedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal compensated steps: %w", err)
	}

	dataJSON, err := json.Marshal(state.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	_, err = r.db.Exec(
		ctx,
		query,
		state.ID,
		state.OrderID,
		string(state.Status),
		state.CurrentStep,
		executedStepsJSON,
		compensatedStepsJSON,
		dataJSON,
		state.Error,
		state.CreatedAt,
		state.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create saga state: %w", err)
	}

	return nil
}

// Update updates an existing saga state.
func (r *SagaStateRepository) Update(ctx context.Context, state *entity.SagaState) error {
	query := `
		UPDATE saga_states SET
			status = $2,
			current_step = $3,
			executed_steps = $4,
			compensated_steps = $5,
			data = $6,
			error = $7,
			updated_at = $8,
			completed_at = $9
		WHERE id = $1`

	executedStepsJSON, err := json.Marshal(state.ExecutedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal executed steps: %w", err)
	}

	compensatedStepsJSON, err := json.Marshal(state.CompensatedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal compensated steps: %w", err)
	}

	dataJSON, err := json.Marshal(state.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	result, err := r.db.Exec(
		ctx,
		query,
		state.ID,
		string(state.Status),
		state.CurrentStep,
		executedStepsJSON,
		compensatedStepsJSON,
		dataJSON,
		state.Error,
		state.UpdatedAt,
		state.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update saga state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("saga state not found: %s", state.ID)
	}

	return nil
}

// FindByID finds a saga state by ID.
func (r *SagaStateRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.SagaState, error) {
	query := `
		SELECT 
			id, order_id, status, current_step,
			executed_steps, compensated_steps, data, error,
			created_at, updated_at, completed_at
		FROM saga_states
		WHERE id = $1`

	row := r.db.QueryRow(ctx, query, id)

	return scanSagaState(row)
}

// FindByOrderID finds a saga state by order ID.
func (r *SagaStateRepository) FindByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*entity.SagaState, error) {
	query := `
		SELECT 
			id, order_id, status, current_step,
			executed_steps, compensated_steps, data, error,
			created_at, updated_at, completed_at
		FROM saga_states
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	row := r.db.QueryRow(ctx, query, orderID)

	return scanSagaState(row)
}

// FindPendingOrFailed finds pending or failed sagas for recovery.
func (r *SagaStateRepository) FindPendingOrFailed(
	ctx context.Context,
	limit int64,
) ([]*entity.SagaState, error) {
	query := `
		SELECT 
			id, order_id, status, current_step,
			executed_steps, compensated_steps, data, error,
			created_at, updated_at, completed_at
		FROM saga_states
		WHERE status IN ($1, $2, $3, $4)
		AND updated_at < NOW() - INTERVAL '1 minute'
		ORDER BY updated_at ASC
		LIMIT $5`

	rows, err := r.db.Query(
		ctx,
		query,
		string(constant.SagaStatusPending),
		string(constant.SagaStatusExecuting),
		string(constant.SagaStatusFailed),
		string(constant.SagaStatusCompensating),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending/failed sagas: %w", err)
	}
	defer rows.Close()

	var states []*entity.SagaState

	for rows.Next() {
		state, err := scanSagaState(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan saga state: %w", err)
		}

		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating saga states: %w", err)
	}

	return states, nil
}

// MarkAsExecuting updates saga status to executing.
func (r *SagaStateRepository) MarkAsExecuting(ctx context.Context, id uuid.UUID) error {
	return r.updateStatus(ctx, id, constant.SagaStatusExecuting)
}

// MarkAsCompensating updates saga status to compensating.
func (r *SagaStateRepository) MarkAsCompensating(ctx context.Context, id uuid.UUID) error {
	return r.updateStatus(ctx, id, constant.SagaStatusCompensating)
}

// MarkAsCompleted updates saga status to completed.
func (r *SagaStateRepository) MarkAsCompleted(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE saga_states 
		SET status = $2, completed_at = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, string(constant.SagaStatusCompleted), time.Now())
	if err != nil {
		return fmt.Errorf("failed to mark saga as completed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("saga state not found: %s", id)
	}

	return nil
}

// MarkAsFailed updates saga status to failed with error message.
func (r *SagaStateRepository) MarkAsFailed(
	ctx context.Context,
	id uuid.UUID,
	errorMsg string,
) error {
	query := `
		UPDATE saga_states 
		SET status = $2, error = $3, completed_at = $4
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		id,
		string(constant.SagaStatusFailed),
		errorMsg,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to mark saga as failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("saga state not found: %s", id)
	}

	return nil
}

// MarkAsCompensated updates saga status to compensated.
func (r *SagaStateRepository) MarkAsCompensated(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE saga_states 
		SET status = $2, completed_at = $3
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, string(constant.SagaStatusCompensated), time.Now())
	if err != nil {
		return fmt.Errorf("failed to mark saga as compensated: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("saga state not found: %s", id)
	}

	return nil
}

// updateStatus updates the status of a saga state.
func (r *SagaStateRepository) updateStatus(
	ctx context.Context,
	id uuid.UUID,
	status constant.SagaStatus,
) error {
	query := `
		UPDATE saga_states 
		SET status = $2
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, string(status))
	if err != nil {
		return fmt.Errorf("failed to update saga status to %s: %w", status, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("saga state not found: %s", id)
	}

	return nil
}

// scanSagaState scans a database row into a SagaState struct.
func scanSagaState(row pgx.Row) (*entity.SagaState, error) {
	var state entity.SagaState

	var statusStr string

	var executedStepsJSON, compensatedStepsJSON, dataJSON []byte

	// Use nullable types for columns that can be NULL
	var errorStr pgtype.Text

	var completedAt pgtype.Timestamptz

	err := row.Scan(
		&state.ID,
		&state.OrderID,
		&statusStr,
		&state.CurrentStep,
		&executedStepsJSON,
		&compensatedStepsJSON,
		&dataJSON,
		&errorStr,
		&state.CreatedAt,
		&state.UpdatedAt,
		&completedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	state.Status = constant.SagaStatus(statusStr)

	// Unmarshal JSON fields
	if err := json.Unmarshal(executedStepsJSON, &state.ExecutedSteps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal executed steps: %w", err)
	}

	if err := json.Unmarshal(compensatedStepsJSON, &state.CompensatedSteps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal compensated steps: %w", err)
	}

	if err := json.Unmarshal(dataJSON, &state.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Handle nullable error
	if errorStr.Status == pgtype.Present {
		state.Error = errorStr.String
	} else {
		state.Error = ""
	}

	// Handle nullable completed_at
	if completedAt.Status == pgtype.Present {
		state.CompletedAt = &completedAt.Time
	} else {
		state.CompletedAt = nil
	}

	return &state, nil
}
