package saga

import (
	"context"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
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
