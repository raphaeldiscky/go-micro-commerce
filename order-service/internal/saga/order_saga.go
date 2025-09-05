// Package saga provides OrderSaga workflow implementation.
package saga

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
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
	// Step 1: Reserve products
	executor.AddStep(&Step{
		Name:        constant.ReserveProductsStep,
		Description: "Reserve products",
		MaxRetries:  3,
		RetryDelay:  2 * time.Second,
		Timeout:     30 * time.Second,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, _ map[string]interface{}) (*StepResult, error) {
			newOrder, reservedProducts, err := s.activities.ReserveProductsAndCalculate(
				ctx.Context(),
				order,
			)
			if err != nil {
				return nil, err
			}

			// Update the original order with new order items from the reservation, save in memory
			order.Items = newOrder.Items

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"reserved_products": reservedProducts,
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) error {
			return s.activities.ReleaseProducts(ctx.Context(), order)
		},
	})

	// Step 2: Create Shipping or Fulfillment
	executor.AddStep(&Step{
		Name:        constant.ProcessFulfillmentStep,
		Description: "Process fulfillment for the order; get shippingID, shippingCost and trackingNumber",
		MaxRetries:  2,
		RetryDelay:  3 * time.Second,
		Timeout:     30 * time.Second,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			shippingID, shippingCost, trackingNumber, err := s.activities.ProcessFulfillment(
				ctx.Context(),
				order,
			)
			if err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"shipping_id":     shippingID,
					"tracking_number": trackingNumber,
					"shipping_cost":   shippingCost,
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

	// Step 3: Set final order prices
	executor.AddStep(&Step{
		Name:        constant.SetFinalPricesStep,
		Description: "Update final order prices to include shipping cost and save to database",
		MaxRetries:  3,
		RetryDelay:  1 * time.Second,
		Timeout:     10 * time.Second,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			shippingCost, ok := data["shipping_cost"].(decimal.Decimal)
			if !ok {
				return nil, fmt.Errorf("shipping cost not found in data")
			}
			// Update the original order with shipping cost

			err := order.UpdateShippingCost(shippingCost)
			if err != nil {
				return nil, err
			}

			if err := s.activities.SetFinalOrderPrices(ctx.Context(), order); err != nil {
				return nil, err
			}

			s.logger.Infof("===STEP 3 COMPLETED===")

			return &StepResult{Success: true}, nil
		},
		Compensate: func(ctx *WorkflowContext, _ *entity.Order, _ map[string]interface{}) error {
			return nil
		},
	})

	// Step 4: Process Payment
	executor.AddStep(&Step{
		Name:        constant.ProcessPaymentStep,
		Description: "Process payment for the order; get payment ID and how much need to be paid",
		MaxRetries:  3,
		RetryDelay:  5 * time.Second,
		Timeout:     60 * time.Second,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, _ map[string]interface{}) (*StepResult, error) {
			s.logger.Infof("===STEP 4===, order: %v", order)
			paymentID, err := s.activities.ProcessPayment(ctx.Context(), order)
			if err != nil {
				return nil, err
			}

			s.logger.Infof("===STEP 4 COMPLETED===, order: %v, paymentID: %s", order, paymentID)

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"payment_id": paymentID,
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) error {
			paymentIDStr, ok := data["payment_id"].(string)
			if !ok {
				return fmt.Errorf("no payment ID found for refund")
			}

			paymentID, err := uuid.Parse(paymentIDStr)
			if err != nil {
				return fmt.Errorf("failed to parse payment ID: %w", err)
			}

			return s.activities.RefundPayment(ctx.Context(), order, paymentID)
		},
	})

	// Step 5: Deduct Products Stock
	executor.AddStep(&Step{
		Name:        constant.ConfirmProductsDeductionStep,
		Description: "Confirms reserved products and release reserved quantity after payment completed",
		MaxRetries:  3,
		RetryDelay:  2 * time.Second,
		Timeout:     20 * time.Second,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			reservedProducts, ok := data["reserved_products"].([]entity.Product)
			if !ok {
				return nil, fmt.Errorf("no reserved products found")
			}
			s.logger.Infof("===STEP 5===, order: %v", order)
			if err := s.activities.ConfirmProductsDeduction(ctx.Context(), order, reservedProducts); err != nil {
				return nil, err
			}

			s.logger.Infof("===STEP 5 COMPLETED===")

			return &StepResult{Success: true}, nil
		},
		Compensate: func(ctx *WorkflowContext, order *entity.Order, _ map[string]interface{}) error {
			return s.activities.RestoreProducts(ctx.Context(), order)
		},
	})

	// Step 6: Send Order Confirmation Notifications
	executor.AddStep(&Step{
		Name:        constant.SendOrderConfirmationStep,
		Description: "Send order confirmation to customer; with invoice need to be paid and tracking number info",
		MaxRetries:  3,
		RetryDelay:  1 * time.Second,
		Timeout:     10 * time.Second,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, order *entity.Order, data map[string]interface{}) (*StepResult, error) {
			s.logger.Infof("===STEP 6===, order: %v", order)
			s.logger.Infof("===DATA===, data: %v", data)
			trackingNumber, ok := data["tracking_number"].(string)
			if !ok {
				ctx.logger.Warn("No tracking number found for notification")

				return nil, fmt.Errorf("no tracking number found")
			}

			if err := s.activities.SendOrderConfirmation(ctx.Context(), order, trackingNumber); err != nil {
				// Non-critical step, log but don't fail the saga
				ctx.logger.Warnf("Failed to send notification: %v", err)
			}

			s.logger.Infof("===STEP 6 COMPLETED===")

			return &StepResult{Success: true}, nil
		},
		Compensate: nil, // No compensation for notifications
	})
}
