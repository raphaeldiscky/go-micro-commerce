package saga

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
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

// GetXEmail returns the X-Email header from the context.
func (wc *WorkflowContext) GetXEmail() (string, error) {
	email, ok := wc.ctx.Value(pkgconstant.CtxEmail).(string)
	if !ok {
		return "", fmt.Errorf("X-Email header not found in context")
	}

	return email, nil
}
