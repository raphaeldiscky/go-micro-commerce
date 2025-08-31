// Package saga provides OrderSaga workflow implementation.
package saga

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// OrderSaga implements the order processing saga workflow.
type OrderSaga struct {
	activities OrderActivities
	logger     logger.Logger
}

// NewOrderSaga creates a new order saga.
func NewOrderSaga(activities OrderActivities, appLogger logger.Logger) *OrderSaga {
	return &OrderSaga{
		activities: activities,
		logger:     appLogger,
	}
}

// ConfigureSteps configures all steps for the order saga.
func (s *OrderSaga) ConfigureSteps(executor *Executor) {
	// Step 1: Validate and Reserve Inventory
	executor.AddStep(Step{
		Name:        "ValidateAndReserveInventory",
		Description: "Validate products and reserve inventory",
		MaxRetries:  3,
		RetryDelay:  2 * time.Second,
		Idempotent:  true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			// Validate products exist and have sufficient stock
			if err := s.activities.ValidateProducts(ctx.Context(), order); err != nil {
				return nil, err
			}

			// Reserve inventory
			reservationID, err := s.activities.ReserveInventory(ctx.Context(), order)
			if err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"reservation_id": reservationID,
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) error {
			reservationID, ok := data["reservation_id"].(uuid.UUID)
			if !ok {
				ctx.logger.Warn("No reservation ID found for compensation")

				return nil
			}

			return s.activities.ReleaseInventoryReservation(ctx.Context(), order, reservationID)
		},
	})

	// Step 2: Calculate Pricing and Discounts
	executor.AddStep(Step{
		Name:        "CalculatePricing",
		Description: "Calculate final pricing with discounts and taxes",
		MaxRetries:  2,
		RetryDelay:  1 * time.Second,
		Idempotent:  true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			pricing, err := s.activities.CalculatePricing(ctx.Context(), order)
			if err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"total_price":    pricing.TotalPrice,
					"total_discount": pricing.Discount,
					"total_tax":      pricing.Tax,
				},
			}, nil
		},
		Compensate: nil, // No compensation needed for calculation
	})

	// Step 3: Process Payment
	executor.AddStep(Step{
		Name:        "ProcessPayment",
		Description: "Process payment for the order",
		MaxRetries:  3,
		RetryDelay:  5 * time.Second,
		Idempotent:  true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			paymentID, err := s.activities.ProcessPayment(ctx.Context(), order)
			if err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"payment_id": paymentID,
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) error {
			paymentID, ok := data["payment_id"].(uuid.UUID)
			if !ok {
				ctx.logger.Warn("No payment ID found for refund")

				return nil
			}

			return s.activities.RefundPayment(ctx.Context(), order, paymentID)
		},
	})

	// Step 4: Confirm Inventory
	executor.AddStep(Step{
		Name:        "ConfirmInventory",
		Description: "Confirm inventory deduction after payment",
		MaxRetries:  3,
		RetryDelay:  2 * time.Second,
		Idempotent:  true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			reservationID, ok := data["reservation_id"].(uuid.UUID)
			if !ok {
				ctx.logger.Warn("No reservation ID found for inventory confirmation")

				return nil, fmt.Errorf("no reservation ID found")
			}

			if err := s.activities.ConfirmInventoryDeduction(ctx.Context(), order, reservationID); err != nil {
				return nil, err
			}

			return &StepResult{Success: true}, nil
		},
		Compensate: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) error {
			// Restore inventory if needed
			return s.activities.RestoreInventory(ctx.Context(), order)
		},
	})

	// Step 5: Create Shipping
	executor.AddStep(Step{
		Name:        "CreateShipping",
		Description: "Create shipping arrangement",
		MaxRetries:  2,
		RetryDelay:  3 * time.Second,
		Idempotent:  true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			shippingID, trackingNumber, err := s.activities.CreateShipping(ctx.Context(), order)
			if err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"shipping_id":     shippingID,
					"tracking_number": trackingNumber,
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) error {
			shippingID, ok := data["shipping_id"].(uuid.UUID)
			if !ok {
				return nil
			}

			return s.activities.CancelShipping(ctx.Context(), shippingID)
		},
	})

	// Step 6: Send Notifications
	executor.AddStep(Step{
		Name:        "SendNotifications",
		Description: "Send order confirmation notifications",
		MaxRetries:  3,
		RetryDelay:  1 * time.Second,
		Idempotent:  true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			trackingNumber, err := data["tracking_number"].(string)
			if !err {
				ctx.logger.Warn("No tracking number found for notification")

				return nil, fmt.Errorf("no tracking number found")
			}

			if err := s.activities.SendOrderConfirmation(ctx.Context(), order, trackingNumber); err != nil {
				// Non-critical step, log but don't fail the saga
				ctx.logger.Warnf("Failed to send notification: %v", err)
			}

			return &StepResult{Success: true}, nil
		},
		Compensate: nil, // No compensation for notifications
	})
}
