package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// SagaStateRepository interface for persisting saga state.
type SagaStateRepository interface {
	// Create inserts a new saga state.
	Create(ctx context.Context, state *entity.SagaState) error
	// Update updates an existing saga state.
	Update(ctx context.Context, state *entity.SagaState) error
	// FindByID finds a saga state by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.SagaState, error)
	// FindByOrderID finds a saga state by order ID (returns most recent).
	FindByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.SagaState, error)
	// FindByOrderIDAndWorkflow finds a saga state by order ID and workflow name.
	FindByOrderIDAndWorkflow(
		ctx context.Context,
		orderID uuid.UUID,
		workflowName constant.WorkflowName,
	) (*entity.SagaState, error)
	// FindPendingOrFailed finds pending or failed sagas for recovery.
	FindPendingOrFailed(ctx context.Context, limit int64) ([]*entity.SagaState, error)
	// FindTimeoutSagas finds sagas that have timed out.
	FindTimeoutSagas(ctx context.Context, limit int64) ([]*entity.SagaState, error)
	// UpdateWithVersion updates saga state with optimistic locking.
	UpdateWithVersion(ctx context.Context, state *entity.SagaState) error
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

// sagaStateRepository implements the SagaStateRepository.
type sagaStateRepository struct {
	db DBTX
}

// NewSagaStateRepository creates a new instance of sagaStateRepository.
func NewSagaStateRepository(db DBTX) SagaStateRepository {
	return &sagaStateRepository{
		db: db,
	}
}

// Create inserts a new saga state.
func (r *sagaStateRepository) Create(ctx context.Context, state *entity.SagaState) error {
	query := `
		INSERT INTO saga_states (
			id, workflow_name, order_id, status, current_step,
			executed_steps, compensated_steps, data, error,
			version, retry_count, timeout_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	executedStepsJSON, err := sonic.Marshal(state.ExecutedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal executed steps: %w", err)
	}

	compensatedStepsJSON, err := sonic.Marshal(state.CompensatedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal compensated steps: %w", err)
	}

	dataJSON, err := sonic.Marshal(state.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	_, err = r.db.Exec(
		ctx,
		query,
		state.ID,
		state.WorkflowName,
		state.OrderID,
		string(state.Status),
		state.CurrentStep,
		executedStepsJSON,
		compensatedStepsJSON,
		dataJSON,
		state.Error,
		state.Version,
		state.RetryCount,
		state.TimeoutAt,
		state.CreatedAt,
		state.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create saga state: %w", err)
	}

	return nil
}

// Update updates an existing saga state.
func (r *sagaStateRepository) Update(ctx context.Context, state *entity.SagaState) error {
	query := `
		UPDATE saga_states SET
			workflow_name = $2,
			status = $3,
			current_step = $4,
			executed_steps = $5,
			compensated_steps = $6,
			data = $7,
			error = $8,
			version = version + 1,
			retry_count = $9,
			last_retry_at = $10,
			updated_at = $11,
			completed_at = $12
		WHERE id = $1`

	executedStepsJSON, err := sonic.Marshal(state.ExecutedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal executed steps: %w", err)
	}

	compensatedStepsJSON, err := sonic.Marshal(state.CompensatedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal compensated steps: %w", err)
	}

	dataJSON, err := sonic.Marshal(state.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	result, err := r.db.Exec(
		ctx,
		query,
		state.ID,
		state.WorkflowName,
		string(state.Status),
		state.CurrentStep,
		executedStepsJSON,
		compensatedStepsJSON,
		dataJSON,
		state.Error,
		state.RetryCount,
		state.LastRetryAt,
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
func (r *sagaStateRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.SagaState, error) {
	query := `
		SELECT
			id, workflow_name, order_id, status, current_step,
			executed_steps, compensated_steps, data, error,
			version, retry_count, last_retry_at, timeout_at,
			created_at, updated_at, completed_at
		FROM saga_states
		WHERE id = $1`

	row := r.db.QueryRow(ctx, query, id)

	return scanSagaState(row)
}

// FindByOrderID finds a saga state by order ID (returns most recent).
func (r *sagaStateRepository) FindByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*entity.SagaState, error) {
	query := `
		SELECT
			id, workflow_name, order_id, status, current_step,
			executed_steps, compensated_steps, data, error,
			version, retry_count, last_retry_at, timeout_at,
			created_at, updated_at, completed_at
		FROM saga_states
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	row := r.db.QueryRow(ctx, query, orderID)

	return scanSagaState(row)
}

// FindByOrderIDAndWorkflow finds a saga state by order ID and workflow name.
func (r *sagaStateRepository) FindByOrderIDAndWorkflow(
	ctx context.Context,
	orderID uuid.UUID,
	workflowName constant.WorkflowName,
) (*entity.SagaState, error) {
	query := `
		SELECT
			id, workflow_name, order_id, status, current_step,
			executed_steps, compensated_steps, data, error,
			version, retry_count, last_retry_at, timeout_at,
			created_at, updated_at, completed_at
		FROM saga_states
		WHERE order_id = $1 AND workflow_name = $2
		ORDER BY created_at DESC
		LIMIT 1`

	row := r.db.QueryRow(ctx, query, orderID, workflowName)

	return scanSagaState(row)
}

// FindPendingOrFailed finds pending or failed sagas for recovery.
func (r *sagaStateRepository) FindPendingOrFailed(
	ctx context.Context,
	limit int64,
) ([]*entity.SagaState, error) {
	query := `
		SELECT
			id, workflow_name, order_id, status, current_step,
			executed_steps, compensated_steps, data, error,
			version, retry_count, last_retry_at, timeout_at,
			created_at, updated_at, completed_at
		FROM saga_states
		WHERE status IN ($1, $2, $3, $4)
		AND updated_at < NOW() - INTERVAL '5 minutes'
		AND (timeout_at IS NULL OR timeout_at > NOW())
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
		state, errNext := scanSagaState(rows)
		if errNext != nil {
			return nil, fmt.Errorf("failed to scan saga state: %w", errNext)
		}

		states = append(states, state)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating saga states: %w", err)
	}

	return states, nil
}

// FindTimeoutSagas finds sagas that have timed out.
func (r *sagaStateRepository) FindTimeoutSagas(
	ctx context.Context,
	limit int64,
) ([]*entity.SagaState, error) {
	query := `
		SELECT
			id, workflow_name, order_id, status, current_step,
			executed_steps, compensated_steps, data, error,
			version, retry_count, last_retry_at, timeout_at,
			created_at, updated_at, completed_at
		FROM saga_states
		WHERE timeout_at IS NOT NULL
		AND timeout_at <= NOW()
		AND status IN ($1, $2, $3)
		ORDER BY timeout_at ASC
		LIMIT $4`

	rows, err := r.db.Query(
		ctx,
		query,
		string(constant.SagaStatusPending),
		string(constant.SagaStatusExecuting),
		string(constant.SagaStatusCompensating),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find timeout sagas: %w", err)
	}
	defer rows.Close()

	var states []*entity.SagaState

	for rows.Next() {
		state, rowErr := scanSagaState(rows)
		if rowErr != nil {
			return nil, fmt.Errorf("failed to scan saga state: %w", rowErr)
		}

		states = append(states, state)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating timeout sagas: %w", err)
	}

	return states, nil
}

// UpdateWithVersion updates saga state with optimistic locking.
func (r *sagaStateRepository) UpdateWithVersion(
	ctx context.Context,
	state *entity.SagaState,
) error {
	executedStepsJSON, err := sonic.Marshal(state.ExecutedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal executed steps: %w", err)
	}

	compensatedStepsJSON, err := sonic.Marshal(state.CompensatedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal compensated steps: %w", err)
	}

	dataJSON, err := sonic.Marshal(state.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	query := `
		UPDATE saga_states SET
			workflow_name = $2,
			status = $3,
			current_step = $4,
			executed_steps = $5,
			compensated_steps = $6,
			data = $7,
			error = $8,
			version = $9 + 1,
			retry_count = $10,
			last_retry_at = $11,
			updated_at = $12,
			completed_at = $13
		WHERE id = $1 AND version = $9`

	result, err := r.db.Exec(
		ctx,
		query,
		state.ID,
		state.WorkflowName,
		string(state.Status),
		state.CurrentStep,
		executedStepsJSON,
		compensatedStepsJSON,
		dataJSON,
		state.Error,
		state.Version,
		state.RetryCount,
		state.LastRetryAt,
		state.UpdatedAt,
		state.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update saga state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("saga state version conflict or not found: %s", state.ID)
	}

	// Increment version in memory
	state.Version++

	return nil
}

// MarkAsExecuting updates saga status to executing.
func (r *sagaStateRepository) MarkAsExecuting(ctx context.Context, id uuid.UUID) error {
	return r.updateStatus(ctx, id, constant.SagaStatusExecuting)
}

// MarkAsCompensating updates saga status to compensating.
func (r *sagaStateRepository) MarkAsCompensating(ctx context.Context, id uuid.UUID) error {
	return r.updateStatus(ctx, id, constant.SagaStatusCompensating)
}

// MarkAsCompleted updates saga status to completed.
func (r *sagaStateRepository) MarkAsCompleted(ctx context.Context, id uuid.UUID) error {
	const query = `
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
func (r *sagaStateRepository) MarkAsFailed(
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
func (r *sagaStateRepository) MarkAsCompensated(ctx context.Context, id uuid.UUID) error {
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
func (r *sagaStateRepository) updateStatus(
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

	var lastRetryAt pgtype.Timestamptz

	var timeoutAt pgtype.Timestamptz

	var completedAt pgtype.Timestamptz

	err := row.Scan(
		&state.ID,
		&state.WorkflowName,
		&state.OrderID,
		&statusStr,
		&state.CurrentStep,
		&executedStepsJSON,
		&compensatedStepsJSON,
		&dataJSON,
		&errorStr,
		&state.Version,
		&state.RetryCount,
		&lastRetryAt,
		&timeoutAt,
		&state.CreatedAt,
		&state.UpdatedAt,
		&completedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.SagaStateNotFoundErrorMessage)
		}

		return nil, err
	}

	state.Status = constant.SagaStatus(statusStr)

	// Unmarshal JSON fields
	if err = sonic.Unmarshal(executedStepsJSON, &state.ExecutedSteps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal executed steps: %w", err)
	}

	if err = sonic.Unmarshal(compensatedStepsJSON, &state.CompensatedSteps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal compensated steps: %w", err)
	}

	if err = sonic.Unmarshal(dataJSON, &state.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Handle nullable error
	if errorStr.Status == pgtype.Present {
		state.Error = errorStr.String
	} else {
		state.Error = ""
	}

	// Handle nullable last_retry_at
	if lastRetryAt.Status == pgtype.Present {
		state.LastRetryAt = &lastRetryAt.Time
	} else {
		state.LastRetryAt = nil
	}

	// Handle nullable timeout_at
	if timeoutAt.Status == pgtype.Present {
		state.TimeoutAt = &timeoutAt.Time
	} else {
		state.TimeoutAt = nil
	}

	// Handle nullable completed_at
	if completedAt.Status == pgtype.Present {
		state.CompletedAt = &completedAt.Time
	} else {
		state.CompletedAt = nil
	}

	return &state, nil
}
