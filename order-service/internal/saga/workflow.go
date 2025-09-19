package saga

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
)

// WorkflowContext provides context for saga execution.
type WorkflowContext struct {
	ctx      context.Context
	orderID  uuid.UUID
	logger   logger.Logger
	userAuth *pkgdto.UserAuthInfo
}

// NewWorkflowContext creates a new workflow context.
func NewWorkflowContext(
	ctx context.Context,
	orderID uuid.UUID,
	appLogger logger.Logger,
	userAuth *pkgdto.UserAuthInfo,
) *WorkflowContext {
	return &WorkflowContext{
		ctx:      ctx,
		orderID:  orderID,
		logger:   appLogger,
		userAuth: userAuth,
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
		return "", errors.New("X-Email header not found in context")
	}

	return email, nil
}

// GetUserAuth returns the user authentication info.
func (wc *WorkflowContext) GetUserAuth() *pkgdto.UserAuthInfo {
	return wc.userAuth
}

// AuthenticatedContext returns a context with proper gRPC authentication headers for external service calls.
func (wc *WorkflowContext) AuthenticatedContext() context.Context {
	// First try to use the stored userAuth
	if wc.userAuth != nil {
		return echoutils.AddUserAuthToContexts(wc.ctx, *wc.userAuth)
	}

	// Fallback: try to extract user auth from context values (for compensation scenarios)
	userAuth, err := echoutils.GetUserAuthContexts(wc.ctx)
	if err == nil {
		return echoutils.AddUserAuthToContexts(wc.ctx, userAuth)
	}

	return wc.ctx
}
