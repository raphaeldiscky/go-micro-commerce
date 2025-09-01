package saga

import (
	"context"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// WorkflowStep represents a step in the workflow.
type WorkflowStep string

const (
	// ValidateProductsStep validates the products in the order.
	ValidateProductsStep WorkflowStep = "ValidateProducts"
	// ReserveProductsStep reserves the products in the order.
	ReserveProductsStep WorkflowStep = "ReserveProducts"
	// CalculatePricingStep calculates the pricing for the order.
	CalculatePricingStep WorkflowStep = "CalculatePricing"
	// ProcessPaymentStep processes the payment for the order.
	ProcessPaymentStep WorkflowStep = "ProcessPayment"
	// ConfirmProductsDeductionStep deducts the products from inventory.
	ConfirmProductsDeductionStep WorkflowStep = "ConfirmProductsDeduction"
	// CreateShippingStep creates a shipping order.
	CreateShippingStep WorkflowStep = "CreateShipping"
	// SendOrderConfirmationStep sends an order confirmation.
	SendOrderConfirmationStep WorkflowStep = "SendOrderConfirmation"
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
