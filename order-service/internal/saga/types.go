package saga

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// WorkflowContext provides context for saga execution.
type WorkflowContext struct {
	ctx     context.Context
	orderID uuid.UUID
	logger  logger.Logger
}

// NewWorkflowContext creates a new workflow context.
func NewWorkflowContext(
	ctx context.Context,
	orderID uuid.UUID,
	appLogger logger.Logger,
) *WorkflowContext {
	return &WorkflowContext{
		ctx:     ctx,
		orderID: orderID,
		logger:  appLogger,
	}
}

// Context returns the underlying context.
func (wc *WorkflowContext) Context() context.Context {
	return wc.ctx
}

// OrderID returns the order ID for this workflow.
func (wc *WorkflowContext) OrderID() uuid.UUID {
	return wc.orderID
}

// Step represents a single step in the saga workflow.
type Step struct {
	Name          string
	Execute       func(ctx *WorkflowContext, order *entity.Order) error
	Compensate    func(ctx *WorkflowContext, order *entity.Order) error
	Description   string
	IsExecuted    bool
	IsCompensated bool
}

// Error represents an error that occurred during saga execution.
type Error struct {
	Step    string
	Message string
	Err     error
}

// Error implements the error interface for SagaError.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("saga step '%s' failed: %s - %v", e.Step, e.Message, e.Err)
	}

	return fmt.Sprintf("saga step '%s' failed: %s", e.Step, e.Message)
}

// Executor handles the execution of saga workflows.
type Executor struct {
	steps  []Step
	logger logger.Logger
}

// NewSagaExecutor creates a new saga executor.
func NewSagaExecutor(appLogger logger.Logger) *Executor {
	return &Executor{
		steps:  make([]Step, 0),
		logger: appLogger,
	}
}

// AddStep adds a step to the saga workflow.
func (se *Executor) AddStep(step Step) {
	se.steps = append(se.steps, step)
}

// Execute runs the saga workflow with compensation on failure.
func (se *Executor) Execute(ctx *WorkflowContext, order *entity.Order) error {
	executedSteps := make([]int, 0)

	// Execute steps in order
	for i, step := range se.steps {
		se.logger.Infof("Executing saga step: %s - %s", step.Name, step.Description)

		if err := step.Execute(ctx, order); err != nil {
			sagaErr := &Error{
				Step:    step.Name,
				Message: "execution failed",
				Err:     err,
			}

			se.logger.Errorf("Saga step failed: %v", sagaErr)

			// Compensate in reverse order
			if compensateErr := se.compensate(ctx, order, executedSteps); compensateErr != nil {
				se.logger.Errorf("Compensation failed: %v", compensateErr)
				// Return both errors
				return fmt.Errorf(
					"saga execution failed: %w, compensation failed: %w",
					sagaErr,
					compensateErr,
				)
			}

			return sagaErr
		}

		se.steps[i].IsExecuted = true

		executedSteps = append(executedSteps, i)

		se.logger.Infof("Successfully executed saga step: %s", step.Name)
	}

	se.logger.Infof("Saga completed successfully for order: %s", order.ID)

	return nil
}

// compensate executes compensation functions in reverse order.
func (se *Executor) compensate(
	ctx *WorkflowContext,
	order *entity.Order,
	executedSteps []int,
) error {
	se.logger.Infof("Starting compensation for %d executed steps", len(executedSteps))

	// Compensate in reverse order
	for i := len(executedSteps) - 1; i >= 0; i-- {
		stepIndex := executedSteps[i]
		step := se.steps[stepIndex]

		if step.Compensate == nil {
			se.logger.Warnf("No compensation function for step: %s", step.Name)

			continue
		}

		se.logger.Infof("Compensating saga step: %s", step.Name)

		if err := step.Compensate(ctx, order); err != nil {
			se.logger.Errorf("Compensation failed for step %s: %v", step.Name, err)

			return &Error{
				Step:    step.Name,
				Message: "compensation failed",
				Err:     err,
			}
		}

		se.steps[stepIndex].IsCompensated = true
		se.logger.Infof("Successfully compensated saga step: %s", step.Name)
	}

	se.logger.Infof("Compensation completed successfully")

	return nil
}

// GetStepStatus returns the status of all steps.
func (se *Executor) GetStepStatus() []Step {
	return se.steps
}
