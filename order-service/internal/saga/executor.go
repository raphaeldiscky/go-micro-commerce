package saga

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
)

// StepResult represents the result of a step execution.
type StepResult struct {
	Success bool
	Data    map[string]interface{}
	Error   error
}

// Step represents an enhanced saga step with retry logic.
type Step struct {
	Name        WorkflowStep
	Execute     func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error)
	Compensate  func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) error
	MaxRetries  int64
	RetryDelay  time.Duration
	Timeout     time.Duration // Individual step timeout
	Description string
	Idempotent  bool // Whether this step is idempotent
	Critical    bool // Whether failure of this step should fail the entire saga
}

// Executor handles saga execution with state persistence and retry logic.
type Executor struct {
	steps      []Step
	dataStore  repository.DataStore
	logger     logger.Logger
	maxRetries int64
	retryDelay time.Duration
}

// NewExecutor creates a new saga executor.
func NewExecutor(
	dataStore repository.DataStore,
	appLogger logger.Logger,
) *Executor {
	return &Executor{
		steps:      make([]Step, 0),
		dataStore:  dataStore,
		logger:     appLogger,
		maxRetries: 3,
		retryDelay: 2 * time.Second,
	}
}

// AddStep adds a step to the saga workflow.
func (e *Executor) AddStep(step *Step) {
	// Set default retry values if not specified
	if step.MaxRetries == 0 {
		step.MaxRetries = e.maxRetries
	}

	if step.RetryDelay == 0 {
		step.RetryDelay = e.retryDelay
	}

	e.steps = append(e.steps, *step)
}

// Execute runs the saga workflow with state persistence and compensation.
func (e *Executor) Execute(ctx context.Context, order *entity.Order) error {
	sagaState, err := e.initializeSagaState(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("failed to initialize saga state: %w", err)
	}

	if sagaState.Status == constant.SagaStatusCompensating {
		return e.compensateFromState(ctx, order, sagaState)
	}

	if err := e.markSagaAsExecuting(ctx, sagaState); err != nil {
		return fmt.Errorf("failed to update saga state: %w", err)
	}

	if err := e.executeAllSteps(ctx, order, sagaState); err != nil {
		return err
	}

	return e.markSagaAsCompleted(ctx, order, sagaState)
}

// initializeSagaState creates or retrieves saga state and sets timeout.
func (e *Executor) initializeSagaState(
	ctx context.Context,
	orderID uuid.UUID,
) (*entity.SagaState, error) {
	sagaState, err := e.getOrCreateSagaState(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Set timeout if not already set
	if sagaState.TimeoutAt == nil {
		sagaState.SetTimeout(30 * time.Minute) // Default saga timeout
	}

	return sagaState, nil
}

// markSagaAsExecuting updates the saga status to executing.
func (e *Executor) markSagaAsExecuting(ctx context.Context, sagaState *entity.SagaState) error {
	stateRepo := e.dataStore.SagaStateRepository()
	sagaState.Status = constant.SagaStatusExecuting
	sagaState.UpdatedAt = time.Now().UTC()

	return stateRepo.UpdateWithVersion(ctx, sagaState)
}

// executeAllSteps executes all saga steps with error handling.
func (e *Executor) executeAllSteps(
	ctx context.Context,
	order *entity.Order,
	sagaState *entity.SagaState,
) error {
	startStep := sagaState.CurrentStep

	for i := startStep; i < int64(len(e.steps)); i++ {
		step := e.steps[i]

		if e.shouldSkipStep(sagaState, &step) {
			continue
		}

		e.logger.Infof("Executing saga step: %s - %s", step.Name, step.Description)

		if err := e.executeSingleStep(ctx, order, &step, sagaState, i); err != nil {
			return err
		}

		e.logger.Infof("Successfully executed saga step: %s", step.Name)
	}

	return nil
}

// shouldSkipStep checks if a step should be skipped.
func (e *Executor) shouldSkipStep(sagaState *entity.SagaState, step *Step) bool {
	if e.isStepExecuted(sagaState, step.Name) && step.Idempotent {
		e.logger.Infof("Step %s already executed, skipping", step.Name)

		return true
	}

	return false
}

// executeSingleStep executes a single step and handles errors.
func (e *Executor) executeSingleStep(
	ctx context.Context,
	order *entity.Order,
	step *Step,
	sagaState *entity.SagaState,
	stepIndex int64,
) error {
	result, err := e.executeStepWithRetry(ctx, order, step, sagaState)
	if err != nil {
		return e.handleStepError(ctx, order, step, sagaState, err)
	}

	return e.updateSagaStateAfterSuccess(ctx, step, sagaState, result, stepIndex)
}

// handleStepError handles step execution errors and compensation.
func (e *Executor) handleStepError(
	ctx context.Context,
	order *entity.Order,
	step *Step,
	sagaState *entity.SagaState,
	err error,
) error {
	e.logger.Errorf("Step %s failed after retries: %v", step.Name, err)

	sagaErr := CategorizeError(step.Name, err)

	// For non-critical steps that are not retriable, continue with warning
	if !step.Critical && !sagaErr.IsRetriable() {
		e.logger.Warnf("Non-critical step %s failed, continuing saga: %v", step.Name, err)

		return nil
	}

	return e.startCompensation(ctx, order, sagaState, err)
}

// startCompensation initiates saga compensation.
func (e *Executor) startCompensation(
	ctx context.Context,
	order *entity.Order,
	sagaState *entity.SagaState,
	originalErr error,
) error {
	stateRepo := e.dataStore.SagaStateRepository()

	// Update saga state to compensating
	sagaState.Status = constant.SagaStatusCompensating
	sagaState.Error = originalErr.Error()
	sagaState.UpdatedAt = time.Now().UTC()

	if updateErr := stateRepo.UpdateWithVersion(ctx, sagaState); updateErr != nil {
		e.logger.Errorf("Failed to update saga state: %v", updateErr)
	}

	// Start compensation
	if compensateErr := e.compensateFromState(ctx, order, sagaState); compensateErr != nil {
		return fmt.Errorf(
			"execution failed: %w, compensation failed: %w",
			originalErr,
			compensateErr,
		)
	}

	return originalErr
}

// updateSagaStateAfterSuccess updates saga state after successful step execution.
func (e *Executor) updateSagaStateAfterSuccess(
	ctx context.Context,
	step *Step,
	sagaState *entity.SagaState,
	result *StepResult,
	stepIndex int64,
) error {
	stateRepo := e.dataStore.SagaStateRepository()

	// Update saga state with successful step
	sagaState.ExecutedSteps = append(sagaState.ExecutedSteps, string(step.Name))
	sagaState.CurrentStep = stepIndex + 1

	if result.Data != nil {
		for k, v := range result.Data {
			sagaState.Data[k] = v
		}
	}

	sagaState.UpdatedAt = time.Now().UTC()

	if err := stateRepo.UpdateWithVersion(ctx, sagaState); err != nil {
		return fmt.Errorf("failed to update saga state after step %s: %w", step.Name, err)
	}

	return nil
}

// markSagaAsCompleted marks the saga as successfully completed.
func (e *Executor) markSagaAsCompleted(
	ctx context.Context,
	order *entity.Order,
	sagaState *entity.SagaState,
) error {
	stateRepo := e.dataStore.SagaStateRepository()

	now := time.Now().UTC()
	sagaState.Status = constant.SagaStatusCompleted
	sagaState.CompletedAt = &now
	sagaState.UpdatedAt = now

	if err := stateRepo.UpdateWithVersion(ctx, sagaState); err != nil {
		return fmt.Errorf("failed to mark saga as completed: %w", err)
	}

	e.logger.Infof("Saga completed successfully for order: %s", order.ID)

	return nil
}

// executeStepWithRetry executes a step with retry logic and timeout.
func (e *Executor) executeStepWithRetry(
	ctx context.Context,
	order *entity.Order,
	step *Step,
	state *entity.SagaState,
) (*StepResult, error) {
	var lastErr error

	var attempt int64
	for attempt = 0; attempt <= step.MaxRetries; attempt++ {
		if attempt > 0 {
			e.logger.Infof("Retrying step %s (attempt %d/%d)", step.Name, attempt, step.MaxRetries)
			time.Sleep(step.RetryDelay * time.Duration(attempt)) // Exponential backoff
		}

		// Use a function literal to scope defer cancel per iteration
		result, err := func() (*StepResult, error) {
			stepCtx := ctx

			var cancel context.CancelFunc

			if step.Timeout > 0 {
				stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
				defer cancel()
			}

			workflowCtx := NewWorkflowContext(stepCtx, order.ID, e.logger)

			return step.Execute(workflowCtx, order, state.Data)
		}()
		if err == nil {
			return result, nil
		}

		lastErr = err
		sagaErr := CategorizeError(step.Name, err)

		e.logger.Warnf("Step %s failed (attempt %d): %v", step.Name, attempt+1, err)

		// Don't retry non-retriable errors
		if !sagaErr.IsRetriable() {
			e.logger.Errorf("Step %s failed with non-retriable error: %v", step.Name, err)

			break
		}

		// Check if context is canceled
		select {
		case <-ctx.Done():
			return nil, NewTimeoutError(step.Name, "context canceled", ctx.Err())
		default:
		}
	}

	return nil, fmt.Errorf(
		"step %s failed after %d attempts: %w",
		step.Name,
		step.MaxRetries+1,
		lastErr,
	)
}

// compensateFromState executes compensation based on saga state.
func (e *Executor) compensateFromState(
	ctx context.Context,
	order *entity.Order,
	state *entity.SagaState,
) error {
	e.logger.Infof("Starting compensation for saga %s", state.ID)
	stateRepo := e.dataStore.SagaStateRepository()

	// Compensate in reverse order
	for i := len(state.ExecutedSteps) - 1; i >= 0; i-- {
		stepName := WorkflowStep(state.ExecutedSteps[i])

		// Skip if already compensated
		if e.isStepCompensated(state, stepName) {
			e.logger.Infof("Step %s already compensated, skipping", stepName)

			continue
		}

		// Find the step
		var step *Step

		for _, s := range e.steps {
			if s.Name == stepName {
				step = &s

				break
			}
		}

		if step == nil {
			e.logger.Errorf("Step %s not found for compensation", stepName)

			continue
		}

		if step.Compensate == nil {
			e.logger.Warnf("No compensation function for step: %s", stepName)

			continue
		}

		e.logger.Infof("Compensating saga step: %s", stepName)

		// Execute compensation with retry
		if err := e.compensateStepWithRetry(ctx, order, step, state); err != nil {
			e.logger.Errorf("Compensation failed for step %s: %v", stepName, err)

			state.Status = constant.SagaStatusFailed
			state.Error = fmt.Sprintf("compensation failed for step %s: %v", stepName, err)
			state.UpdatedAt = time.Now().UTC()

			if updateErr := stateRepo.UpdateWithVersion(ctx, state); updateErr != nil {
				e.logger.Errorf("Failed to update saga state: %v", updateErr)
			}

			return err
		}

		// Update state
		state.CompensatedSteps = append(state.CompensatedSteps, string(stepName))
		state.UpdatedAt = time.Now().UTC()

		if err := stateRepo.UpdateWithVersion(ctx, state); err != nil {
			e.logger.Errorf("Failed to update saga state after compensation: %v", err)
		}

		e.logger.Infof("Successfully compensated saga step: %s", stepName)
	}

	// Mark saga as compensated
	now := time.Now().UTC()
	state.Status = constant.SagaStatusCompensated
	state.CompletedAt = &now
	state.UpdatedAt = now

	if err := stateRepo.UpdateWithVersion(ctx, state); err != nil {
		return fmt.Errorf("failed to mark saga as compensated: %w", err)
	}

	e.logger.Infof("Compensation completed successfully for saga %s", state.ID)

	return nil
}

// compensateStepWithRetry executes compensation with retry logic.
func (e *Executor) compensateStepWithRetry(
	ctx context.Context,
	order *entity.Order,
	step *Step,
	state *entity.SagaState,
) error {
	workflowCtx := NewWorkflowContext(ctx, order.ID, e.logger)

	var lastErr error

	var attempt int64
	for attempt = 0; attempt <= step.MaxRetries; attempt++ {
		if attempt > 0 {
			e.logger.Infof("Retrying compensation for step %s (attempt %d/%d)",
				step.Name, attempt, step.MaxRetries)
			time.Sleep(step.RetryDelay * time.Duration(attempt))
		}

		err := step.Compensate(workflowCtx, order, state.Data)
		if err == nil {
			return nil
		}

		lastErr = err
		e.logger.Warnf("Compensation for step %s failed (attempt %d): %v",
			step.Name, attempt+1, err)
	}

	return fmt.Errorf("compensation for step %s failed after %d attempts: %w",
		step.Name, step.MaxRetries+1, lastErr)
}

// getOrCreateSagaState retrieves existing saga state or creates new one.
func (e *Executor) getOrCreateSagaState(
	ctx context.Context,
	orderID uuid.UUID,
) (*entity.SagaState, error) {
	stateRepo := e.dataStore.SagaStateRepository()
	// Try to find existing saga state
	state, err := stateRepo.FindByOrderID(ctx, orderID)
	if err == nil && state != nil {
		e.logger.Infof("Resuming saga %s for order %s from step %d",
			state.ID, orderID, state.CurrentStep)

		return state, nil
	}

	// Create new saga state
	state = &entity.SagaState{
		ID:               uuid.New(),
		OrderID:          orderID,
		Status:           constant.SagaStatusPending,
		CurrentStep:      0,
		ExecutedSteps:    []string{},
		CompensatedSteps: []string{},
		Data:             make(map[string]interface{}),
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err := stateRepo.Create(ctx, state); err != nil {
		return nil, fmt.Errorf("failed to create saga state: %w", err)
	}

	e.logger.Infof("Created new saga %s for order %s", state.ID, orderID)

	return state, nil
}

// Helper functions.
func (e *Executor) isStepExecuted(state *entity.SagaState, stepName WorkflowStep) bool {
	for _, name := range state.ExecutedSteps {
		if name == string(stepName) {
			return true
		}
	}

	return false
}

func (e *Executor) isStepCompensated(state *entity.SagaState, stepName WorkflowStep) bool {
	for _, name := range state.CompensatedSteps {
		if name == string(stepName) {
			return true
		}
	}

	return false
}
