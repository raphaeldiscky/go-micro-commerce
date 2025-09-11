package temporal

import (
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
)

// OrderSagaWorkflow implements the order processing saga using Temporal.
func OrderSagaWorkflow(
	ctx workflow.Context,
	req dto.TemporalOrderSagaRequest,
	config *config.TemporalConfig,
) (*dto.TemporalOrderSagaResponse, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Order Saga Workflow", "orderID", req.Order.ID)

	// Initialize workflow state
	state := &dto.TemporalWorkflowState{
		OrderID:        req.Order.ID,
		CompletedSteps: make(map[constant.WorkflowStep]bool),
	}

	// Configure activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: config.WorkflowTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    config.RetryInterval,
			BackoffCoefficient: config.BackoffCoefficient,
			MaximumInterval:    config.MaxInterval,
			MaximumAttempts:    config.MaxAttempts,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Execute saga steps with compensation on failure
	if err := executeSagaSteps(ctx, req.Order, state, *req.UserAuth, req.Shipping); err != nil {
		logger.Error(
			"Saga execution failed, starting compensation",
			"error",
			err,
			"orderID",
			req.Order.ID,
		)

		// Execute compensation in reverse order
		executeCompensation(ctx, req.Order, state, *req.UserAuth)

		return &dto.TemporalOrderSagaResponse{
			Success: false,
			OrderID: req.Order.ID,
			Error:   err.Error(),
		}, err
	}

	logger.Info("Order Saga Workflow completed successfully", "orderID", req.Order.ID)

	return &dto.TemporalOrderSagaResponse{
		Success:        true,
		OrderID:        req.Order.ID,
		PaymentID:      state.PaymentID,
		ShippingID:     state.ShippingID,
		TrackingNumber: state.TrackingNumber,
		Pricing:        state.Pricing,
	}, nil
}
