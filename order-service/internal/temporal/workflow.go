package temporal

import (
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
)

// OrderSagaWorkflow implements the order processing saga using Temporal.
//
// This workflow orchestrates the complete order processing flow:
// 1. Product reservation and order calculation
// 2. Shipping cost calculation
// 3. Final price setting with shipping
// 4. Payment creation and processing
// 5. Payment confirmation with reminders
// 6. Fulfillment processing
// 7. Product deduction confirmation
// 8. Order confirmation notification
//
// The workflow supports automatic compensation (rollback) if any critical step fails.
// Non-critical steps (like notifications) can fail without triggering compensation.
//
// Parameters:
//   - ctx: Temporal workflow context for activity execution
//   - req: Order saga request containing order, shipping, and user authentication
//   - config: Temporal configuration for timeouts and retry policies
//
// Returns:
//   - TemporalOrderSagaResponse: Result containing success status, order details, and any errors
//   - error: Non-nil if the workflow fails catastrophically
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
	if err := executeSagaSteps(ctx, req.Order, state, *req.UserAuth); err != nil {
		logger.Error(
			"Saga execution failed, starting compensation",
			"error",
			err,
			"orderID",
			req.Order.ID,
		)

		// Execute compensation in reverse order
		compensationErr := executeCompensation(ctx, req.Order, state, *req.UserAuth)
		if compensationErr != nil {
			logger.Error(
				"Compensation failed, workflow marked as failed",
				"compensationError", compensationErr,
				"originalError", err,
				"orderID", req.Order.ID,
			)

			return &dto.TemporalOrderSagaResponse{
				Success: false,
				OrderID: req.Order.ID,
				Error: fmt.Sprintf(
					"saga failed and compensation failed: %v (original error: %v)",
					compensationErr,
					err,
				),
			}, compensationErr
		}

		logger.Info(
			"Saga failed but compensation completed successfully, workflow marked as completed",
			"originalError", err,
			"orderID", req.Order.ID,
		)

		return &dto.TemporalOrderSagaResponse{
			Success:     true, // Successful compensation means the workflow completed successfully
			OrderID:     req.Order.ID,
			Compensated: true,
			Error: fmt.Sprintf(
				"order processing failed but was compensated successfully: %v",
				err,
			),
		}, nil
	}

	logger.Info("Order Saga Workflow completed successfully", "orderID", req.Order.ID)

	return &dto.TemporalOrderSagaResponse{
		Success:        true,
		OrderID:        req.Order.ID,
		PaymentID:      state.PaymentID,
		ShippingID:     state.ShippingID,
		TrackingNumber: state.TrackingNumber,
		TotalPrice:     state.TotalPrice,
		TotalDiscount:  state.TotalDiscount,
		TotalTax:       state.TotalTax,
	}, nil
}
