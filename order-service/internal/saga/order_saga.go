// Package saga provides OrderSaga workflow implementation.
package saga

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/shopspring/decimal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
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

//nolint:funlen,revive,gocyclo,cyclop,gocognit // ConfigureSteps configures all steps for the order saga.
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
		Execute: func(ctx *WorkflowContext, payload *Payload, _ map[string]interface{}) (*StepResult, error) {
			s.logger.Infof("===STEP 1=====, Reserve products: %+v", payload)
			newOrder, reservedProducts, err := s.activities.ReserveProductsAndCalculate(
				ctx.Context(),
				payload.Order,
			)
			if err != nil {
				return nil, err
			}

			// Update the original order with new order items from the reservation, save in memory
			payload.Order.Items = newOrder.Items

			email, err := ctx.GetXEmail()
			if err != nil {
				return nil, err
			}

			s.logger.Infof("===STEP 1 COMPLETED=====, email: %+v", email)

			s.logger.Infof("===STEP 1 storing shipping data===: %+v", payload.Shipping)

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"reserved_products": reservedProducts,
					"customer_email":    email,
					"shipping":          payload.Shipping, // Store shipping data for recovery and later steps
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, payload *Payload, data map[string]interface{}) error {
			return s.activities.ReleaseProducts(ctx.Context(), payload.Order)
		},
	})

	// Step 2: Get Shipping Cost
	executor.AddStep(&Step{
		Name:        constant.GetShippingCostStep,
		Description: "Get shipping cost from shipping service",
		MaxRetries:  2,
		RetryDelay:  3 * time.Second,
		Timeout:     30 * time.Second,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, payload *Payload, data map[string]interface{}) (*StepResult, error) {
			s.logger.Infof("===STEP 2=====, payload: %+v", payload)
			s.logger.Infof("===STEP 2=====, data: %+v", data)

			shippingData, exists := data["shipping"]
			if !exists {
				return nil, fmt.Errorf("shipping data not found in saga state")
			}

			// Convert map to dto.Shipping using JSON marshal/unmarshal
			shippingBytes, err := json.Marshal(shippingData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal shipping data: %w", err)
			}

			var shipping *dto.Shipping
			if err := json.Unmarshal(shippingBytes, &shipping); err != nil {
				return nil, fmt.Errorf("failed to unmarshal shipping data: %w", err)
			}
			shippingCost, err := s.activities.GetShippingCost(
				ctx.Context(),
				payload.Order,
				shipping,
			)
			if err != nil {
				return nil, err
			}

			s.logger.Infof("===STEP 2 COMPLETED=== shipping cost: %+v", shippingCost)

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"shipping_cost": shippingCost,
				},
			}, nil
		},
		Compensate: nil,
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
		Execute: func(ctx *WorkflowContext, payload *Payload, data map[string]interface{}) (*StepResult, error) {
			shippingCost, ok := data["shipping_cost"].(decimal.Decimal)
			if !ok {
				return nil, fmt.Errorf("shipping cost not found in data")
			}

			err := payload.Order.UpdateShippingCost(shippingCost)
			if err != nil {
				return nil, err
			}

			if err := s.activities.SetFinalOrderPrices(ctx.Context(), payload.Order); err != nil {
				return nil, err
			}

			s.logger.Infof("===STEP 3 COMPLETED===")

			return &StepResult{Success: true}, nil
		},
		Compensate: func(ctx *WorkflowContext, _ *Payload, _ map[string]interface{}) error {
			return nil
		},
	})

	// Step 4: Create Payment
	executor.AddStep(&Step{
		Name:        constant.CreatePaymentStep,
		Description: "Create payment record for the order; get payment ID and how much need to be paid",
		MaxRetries:  3,
		RetryDelay:  5 * time.Second,
		Timeout:     60 * time.Second,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, payload *Payload, _ map[string]interface{}) (*StepResult, error) {
			s.logger.Infof("===STEP 4===, order: %v", payload.Order)
			paymentID, err := s.activities.CreatePayment(ctx.Context(), payload.Order)
			if err != nil {
				return nil, err
			}

			s.logger.Infof(
				"===STEP 4 COMPLETED===, order: %v, paymentID: %s",
				payload.Order,
				paymentID,
			)

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"payment_id": paymentID,
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, payload *Payload, data map[string]interface{}) error {
			paymentIDStr, ok := data["payment_id"].(string)
			if !ok {
				return fmt.Errorf("no payment ID found for refund")
			}

			paymentID, err := uuid.Parse(paymentIDStr)
			if err != nil {
				return fmt.Errorf("failed to parse payment ID: %w", err)
			}

			return s.activities.RefundPayment(ctx.Context(), payload.Order, paymentID)
		},
	})

	// Step 5: Wait for payment confirmation
	executor.AddStep(&Step{
		Name:        constant.WaitForPaymentConfirmationStep,
		Description: "Wait for payment confirmation",
		MaxRetries:  3,
		RetryDelay:  5 * time.Second,
		Timeout:     1 * time.Hour, // 1-hour timeout for user payment confirmation
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, payload *Payload, _ map[string]interface{}) (*StepResult, error) {
			s.logger.Infof("===STEP 5===, order: %v", payload.Order)
			paymentID, err := s.activities.WaitForPaymentConfirmation(ctx.Context(), payload.Order)
			if err != nil {
				return nil, err
			}

			s.logger.Infof(
				"===STEP 5 COMPLETED===, order: %v, paymentID: %s",
				payload.Order,
				paymentID,
			)

			return &StepResult{
				Success: true,
				Data: map[string]interface{}{
					"payment_confirmed_id": paymentID,
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, payload *Payload, _ map[string]interface{}) error {
			return nil
		},
	})

	// Step 6: Create Shipping or Fulfillment
	executor.AddStep(&Step{
		Name:        constant.ProcessFulfillmentStep,
		Description: "Process fulfillment for the paid order; get shippingID, shippingCost and trackingNumber",
		MaxRetries:  2,
		RetryDelay:  3 * time.Second,
		Timeout:     30 * time.Second,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, payload *Payload, data map[string]interface{}) (*StepResult, error) {
			shippingID, shippingCost, trackingNumber, err := s.activities.ProcessFulfillment(
				ctx.Context(),
				payload,
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
		Compensate: func(ctx *WorkflowContext, _ *Payload, data map[string]interface{}) error {
			shippingID, ok := data["shipping_id"].(uuid.UUID)
			if !ok {
				return nil
			}

			return s.activities.CancelShipping(ctx.Context(), shippingID)
		},
	})

	// Step 7: Deduct Products Stock
	executor.AddStep(&Step{
		Name:        constant.ConfirmProductsDeductionStep,
		Description: "Permanently deduct products from inventory after successful payment",
		MaxRetries:  3,
		RetryDelay:  2 * time.Second,
		Timeout:     20 * time.Second,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, payload *Payload, data map[string]interface{}) (*StepResult, error) {
			reservedProducts, ok := data["reserved_products"].([]entity.Product)
			if !ok {
				return nil, fmt.Errorf("no reserved products found")
			}
			s.logger.Infof("===STEP 5===, order: %v", payload.Order)
			if err := s.activities.ConfirmProductsDeduction(ctx.Context(), payload.Order, reservedProducts); err != nil {
				return nil, err
			}

			s.logger.Infof("===STEP 5 COMPLETED===")

			return &StepResult{Success: true}, nil
		},
		Compensate: func(ctx *WorkflowContext, payload *Payload, _ map[string]interface{}) error {
			return s.activities.RestoreProducts(ctx.Context(), payload.Order)
		},
	})

	// Step 8: Send Order Confirmation Notifications
	executor.AddStep(&Step{
		Name:        constant.SendOrderConfirmationStep,
		Description: "Send order confirmation and receipt to customer; includes invoice and tracking info",
		MaxRetries:  3,
		RetryDelay:  1 * time.Second,
		Timeout:     10 * time.Second,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, payload *Payload, data map[string]interface{}) (*StepResult, error) {
			s.logger.Infof("===STEP 6===, order: %v", payload.Order)
			s.logger.Infof("===DATA===, data: %v", data)
			trackingNumber, ok := data["tracking_number"].(string)
			if !ok {
				ctx.logger.Warn("No tracking number found for notification")

				return nil, fmt.Errorf("no tracking number found")
			}

			resevedProducts, ok := data["reserved_products"].([]entity.Product)
			if !ok {
				ctx.logger.Error("No reserved products found for notification")

				return nil, fmt.Errorf("no reserved products found")
			}

			customerEmail, ok := data["customer_email"].(string)
			if !ok {
				ctx.logger.Error("No customer email found for notification")

				return nil, fmt.Errorf("no customer email found")
			}

			if err := s.activities.SendOrderConfirmation(ctx.Context(), payload.Order, resevedProducts, trackingNumber, customerEmail); err != nil {
				// Non-critical step, log but don't fail the saga
				ctx.logger.Warnf("Failed to send notification: %v", err)
			}

			s.logger.Infof("===STEP 6 COMPLETED===")

			return &StepResult{Success: true}, nil
		},
		Compensate: nil, // No compensation for notifications
	})
}
