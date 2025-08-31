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

//nolint:funlen,revive // ConfigureSteps configures all steps for the order saga.
func (s *OrderSaga) ConfigureSteps(executor *Executor) {
	// Step 1: Validate Products
	executor.AddStep(&Step{
		Name:        ValidateProductsStep,
		Description: "Validate products",
		MaxRetries:  3,
		RetryDelay:  2 * time.Second,
		Timeout:     30 * time.Second,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, _ map[string]interface{}) (*StepResult, error) {
			// Validate products exist and have sufficient stock
			if err := s.activities.ValidateProducts(ctx.Context(), order); err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
			}, nil
		},
		Compensate: nil,
	})
	// Step 2: Reserve Products
	executor.AddStep(&Step{
		Name:        ReserveProductsStep,
		Description: "Reserve products",
		MaxRetries:  3,
		RetryDelay:  2 * time.Second,
		Timeout:     30 * time.Second,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, _ map[string]interface{}) (*StepResult, error) {
			// Reserve products
			reservationID, err := s.activities.ReserveProducts(ctx.Context(), order)
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

			return s.activities.ReleaseProducts(ctx.Context(), order, reservationID)
		},
	})

	// Step 3: Calculate Pricing and Discounts
	executor.AddStep(&Step{
		Name:        CalculatePricingStep,
		Description: "Calculate final pricing with discounts and taxes",
		MaxRetries:  2,
		RetryDelay:  1 * time.Second,
		Timeout:     15 * time.Second,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, _ map[string]interface{}) (*StepResult, error) {
			pricing, err := s.activities.CalculatePricing(ctx.Context(), order)
			if err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"total_price":    pricing.TotalPrice,
					"total_discount": pricing.TotalDiscount,
					"total_tax":      pricing.TotalTax,
				},
			}, nil
		},
		Compensate: nil, // No compensation needed for calculation
	})

	// Step 4: Process Payment
	executor.AddStep(&Step{
		Name:        ProcessPaymentStep,
		Description: "Process payment for the order",
		MaxRetries:  3,
		RetryDelay:  5 * time.Second,
		Timeout:     60 * time.Second,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, _ map[string]interface{}) (*StepResult, error) {
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

	// Step 5: Deduct Products
	executor.AddStep(&Step{
		Name:        DeductProductsStep,
		Description: "Deduct products after payment completed",
		MaxRetries:  3,
		RetryDelay:  2 * time.Second,
		Timeout:     20 * time.Second,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			reservationID, ok := data["reservation_id"].(uuid.UUID)
			if !ok {
				ctx.logger.Warn("No reservation ID found for stock confirmation")

				return nil, fmt.Errorf("no reservation ID found")
			}

			if err := s.activities.DeductProducts(ctx.Context(), order, reservationID); err != nil {
				return nil, err
			}

			return &StepResult{Success: true}, nil
		},
		Compensate: func(ctx *WorkflowContext, order *entity.Order, _ map[string]interface{}) error {
			return s.activities.RestoreProducts(ctx.Context(), order)
		},
	})

	// Step 6: Create Shipping
	executor.AddStep(&Step{
		Name:        CreateShippingStep,
		Description: "Create shipping arrangement",
		MaxRetries:  2,
		RetryDelay:  3 * time.Second,
		Timeout:     30 * time.Second,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, order *entity.Order, _ map[string]interface{}) (*StepResult, error) {
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
		Compensate: func(ctx *WorkflowContext, _ *entity.Order, data map[string]interface{}) error {
			shippingID, ok := data["shipping_id"].(uuid.UUID)
			if !ok {
				return nil
			}

			return s.activities.CancelShipping(ctx.Context(), shippingID)
		},
	})

	// Step 7: Send Order Confirmation Notifications
	executor.AddStep(&Step{
		Name:        SendOrderConfirmationStep,
		Description: "Send order confirmation",
		MaxRetries:  3,
		RetryDelay:  1 * time.Second,
		Timeout:     10 * time.Second,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			trackingNumber, ok := data["tracking_number"].(string)
			if !ok {
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
