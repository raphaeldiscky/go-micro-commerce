// Package saga provides OrderSaga workflow implementation.
package saga

import (
	"context"
	"fmt"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// OrderSaga implements the order processing saga workflow.
type OrderSaga struct {
	activities OrderActivities
	logger     logger.Logger
}

// NewOrderSaga creates a new OrderSaga instance.
func NewOrderSaga(activities OrderActivities, appLogger logger.Logger) *OrderSaga {
	return &OrderSaga{
		activities: activities,
		logger:     appLogger,
	}
}

// Execute runs the order saga workflow with compensation logic.
func (s *OrderSaga) Execute(ctx context.Context, order *entity.Order) error {
	workflowCtx := NewWorkflowContext(ctx, order.ID, s.logger)
	s.logger.Infof("OrderSaga started for OrderID: %s", order.ID)

	// Create saga executor
	executor := NewSagaExecutor(s.logger)

	// Add saga steps
	s.addSagaSteps(executor)

	// Execute the saga
	if err := executor.Execute(workflowCtx, order); err != nil {
		s.logger.Errorf("OrderSaga failed for OrderID %s: %v", order.ID, err)

		return fmt.Errorf("order saga failed: %w", err)
	}

	s.logger.Infof("OrderSaga completed successfully for OrderID: %s", order.ID)

	return nil
}

// addSagaSteps configures all the steps for the order saga.
func (s *OrderSaga) addSagaSteps(executor *Executor) {
	// Step 1: Reserve Inventory
	executor.AddStep(Step{
		Name:        "ReserveInventory",
		Description: "Reserve product inventory for the order",
		Execute: func(ctx *WorkflowContext, order *entity.Order) error {
			return s.activities.ReserveInventory(ctx.Context(), order)
		},
		Compensate: func(ctx *WorkflowContext, order *entity.Order) error {
			return s.activities.ReleaseInventoryReservation(ctx.Context(), order)
		},
	})

	// Step 2: Process Payment
	executor.AddStep(Step{
		Name:        "ProcessPayment",
		Description: "Process payment for the order",
		Execute: func(ctx *WorkflowContext, order *entity.Order) error {
			return s.activities.ProcessPayment(ctx.Context(), order)
		},
		Compensate: func(ctx *WorkflowContext, order *entity.Order) error {
			return s.activities.RefundPayment(ctx.Context(), order)
		},
	})

	// Step 3: Update Inventory
	executor.AddStep(Step{
		Name:        "UpdateInventory",
		Description: "Update product inventory after successful payment",
		Execute: func(ctx *WorkflowContext, order *entity.Order) error {
			return s.activities.UpdateInventory(ctx.Context(), order)
		},
		// No compensation needed for this step as inventory is already updated
		Compensate: nil,
	})

	// Step 4: Arrange Shipping
	executor.AddStep(Step{
		Name:        "ArrangeShipping",
		Description: "Arrange shipping for the completed order",
		Execute: func(ctx *WorkflowContext, order *entity.Order) error {
			return s.activities.ArrangeShipping(ctx.Context(), order)
		},
		Compensate: func(_ *WorkflowContext, order *entity.Order) error {
			// In real scenarios, you might want to cancel shipping arrangements
			s.logger.Warnf(
				"Shipping arrangement compensation not implemented for order: %s",
				order.ID,
			)

			return nil
		},
	})
}
