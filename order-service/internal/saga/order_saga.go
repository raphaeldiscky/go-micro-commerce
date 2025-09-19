// Package saga provides OrderSaga workflow implementation.
package saga

import (
	"errors"
	"fmt"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// OrderSaga implements the order processing saga workflow.
type OrderSaga struct {
	activities OrderActivities
}

// NewOrderSaga creates a new order saga.
func NewOrderSaga(activities OrderActivities) *OrderSaga {
	return &OrderSaga{
		activities: activities,
	}
}

//nolint:funlen,revive,gocyclo,cyclop,gocognit // ConfigureSteps configures all steps for the order saga.
func (s *OrderSaga) ConfigureSteps(executor *Executor) {
	// Step 1: Reserve products
	executor.AddStep(&Step{
		Name:        constant.ReserveProductsStep,
		Description: "Reserve products",
		MaxRetries:  constant.ReserveProductsStepMaxRetries,
		RetryDelay:  constant.ReserveProductsStepRetryDelay,
		Timeout:     constant.ReserveProductsStepTimeout,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, payload *Payload, _ *Metadata) (*StepResult, error) {
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

			// Get user auth from context using echoutils
			userAuth, authErr := echoutils.GetUserAuthContexts(ctx.Context())
			if authErr != nil {
				return nil, fmt.Errorf("failed to get user auth: %w", authErr)
			}

			return &StepResult{
				Success: true,
				Data: &Metadata{
					ReservedProducts: reservedProducts,
					CustomerEmail:    email,
					Shipping:         &payload.Shipping, // Store shipping data for recovery and later steps
					UserAuth:         &userAuth,         // Store user auth for all subsequent steps
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, payload *Payload, data *Metadata) error {
			return s.activities.ReleaseProducts(ctx.Context(), payload.Order)
		},
	})

	// Step 2: Get Shipping Cost
	executor.AddStep(&Step{
		Name:        constant.GetShippingCostStep,
		Description: "Get shipping cost from shipping service",
		MaxRetries:  constant.GetShippingCostStepMaxRetries,
		RetryDelay:  constant.GetShippingCostStepRetryDelay,
		Timeout:     constant.GetShippingCostStepTimeout,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, payload *Payload, data *Metadata) (*StepResult, error) {
			if data.Shipping == nil {
				return nil, errors.New("shipping data not found in saga state")
			}

			shippingCost, err := s.activities.GetShippingCost(
				ctx.Context(),
				payload.Order,
				data.Shipping,
			)
			if err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
				Data: &Metadata{
					ShippingCost: &shippingCost,
				},
			}, nil
		},
		Compensate: nil,
	})

	// Step 3: Set final order prices
	executor.AddStep(&Step{
		Name:        constant.SetFinalPricesStep,
		Description: "Update final order prices to include shipping cost and save to database",
		MaxRetries:  constant.SetFinalPricesStepMaxRetries,
		RetryDelay:  constant.SetFinalPricesStepRetryDelay,
		Timeout:     constant.SetFinalPricesStepTimeout,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, payload *Payload, data *Metadata) (*StepResult, error) {
			if data.ShippingCost == nil {
				return nil, errors.New("shipping cost not found in data")
			}

			err := payload.Order.UpdateShippingCost(*data.ShippingCost)
			if err != nil {
				return nil, err
			}

			if err = s.activities.SetFinalOrderPrices(ctx.Context(), payload.Order); err != nil {
				return nil, err
			}

			return &StepResult{Success: true}, nil
		},
		Compensate: func(ctx *WorkflowContext, _ *Payload, _ *Metadata) error {
			return nil
		},
	})

	// Step 4: Create Payment
	executor.AddStep(&Step{
		Name:        constant.CreatePaymentStep,
		Description: "Create payment record for the order; get payment ID and how much need to be paid",
		MaxRetries:  constant.CreatePaymentStepMaxRetries,
		RetryDelay:  constant.CreatePaymentStepRetryDelay,
		Timeout:     constant.CreatePaymentStepTimeout,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, payload *Payload, _ *Metadata) (*StepResult, error) {
			paymentID, err := s.activities.CreatePayment(ctx.Context(), payload.Order)
			if err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
				Data: &Metadata{
					PaymentID: &paymentID,
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, payload *Payload, data *Metadata) error {
			if data.PaymentID == nil {
				// No payment was created, so no refund needed - this is normal for expired orders
				ctx.logger.Infof(
					"No payment ID found for order %s, skipping refund compensation",
					payload.Order.ID,
				)
				return nil
			}

			return s.activities.RefundPayment(
				ctx.Context(),
				payload.Order,
				*data.PaymentID,
			)
		},
	})

	// Step 5: Send Payment Required Notification
	executor.AddStep(&Step{
		Name:        constant.SendPaymentRequiredNotificationStep,
		Description: "Send payment required notification",
		MaxRetries:  constant.SendPaymentRequiredNotificationStepMaxRetries,
		RetryDelay:  constant.SendPaymentRequiredNotificationStepRetryDelay,
		Timeout:     constant.SendPaymentRequiredNotificationStepTimeout,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, payload *Payload, data *Metadata) (*StepResult, error) {
			if len(data.ReservedProducts) == 0 {
				return nil, errors.New("reserved products not found in data")
			}
			if data.CustomerEmail == "" {
				ctx.logger.Error("No customer email found for notification")

				return nil, errors.New("no customer email found")
			}

			err := s.activities.SendPaymentRequiredNotification(
				ctx.Context(),
				payload.Order,
				data.ReservedProducts,
				data.CustomerEmail,
			)
			if err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
				Data:    data,
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, payload *Payload, _ *Metadata) error {
			return nil
		},
	})

	// Step 6: Wait for payment confirmation
	executor.AddStep(&Step{
		Name:        constant.WaitForPaymentConfirmationStep,
		Description: "Wait for customer to pay the order",
		MaxRetries:  constant.WaitForPaymentConfirmationStepMaxRetries,
		RetryDelay:  constant.WaitForPaymentConfirmationStepRetryDelay,
		Timeout:     constant.WaitForPaymentConfirmationStepTimeout, // 5 minutes timeout for user payment confirmation
		Idempotent:  false,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, payload *Payload, data *Metadata) (*StepResult, error) {
			if data.CustomerEmail == "" {
				ctx.logger.Error("No customer email found for notification")
				return nil, errors.New("no customer email found")
			}
			paymentID, err := s.activities.WaitForPaymentConfirmation(
				ctx.Context(),
				payload.Order,
				data.CustomerEmail,
			)
			if err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
				Data: &Metadata{
					PaymentID: &paymentID,
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, payload *Payload, _ *Metadata) error {
			return nil
		},
	})

	// Step 7: Create Shipping or Fulfillment
	executor.AddStep(&Step{
		Name:        constant.ProcessFulfillmentStep,
		Description: "Process fulfillment for the paid order; get fulfillmentID, shippingCost and trackingNumber",
		MaxRetries:  constant.ProcessFulfillmentStepMaxRetries,
		RetryDelay:  constant.ProcessFulfillmentStepRetryDelay,
		Timeout:     constant.ProcessFulfillmentStepTimeout,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, payload *Payload, data *Metadata) (*StepResult, error) {
			fulfillmentID, shippingCost, trackingNumber, err := s.activities.ProcessFulfillment(
				ctx.Context(),
				payload,
			)
			if err != nil {
				return nil, err
			}

			return &StepResult{
				Success: true,
				Data: &Metadata{
					FulfillmentID:  &fulfillmentID,
					TrackingNumber: &trackingNumber,
					ShippingCost:   &shippingCost,
				},
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, _ *Payload, data *Metadata) error {
			if data.FulfillmentID == nil {
				return nil
			}
			return s.activities.CancelShipping(ctx.Context(), *data.FulfillmentID)
		},
	})

	// Step 8: Deduct Products Stock
	executor.AddStep(&Step{
		Name:        constant.ConfirmProductsDeductionStep,
		Description: "Permanently deduct products from inventory after successful payment",
		MaxRetries:  constant.ConfirmProductsDeductionStepMaxRetries,
		RetryDelay:  constant.ConfirmProductsDeductionStepRetryDelay,
		Timeout:     constant.ConfirmProductsDeductionStepTimeout,
		Idempotent:  true,
		Critical:    true,
		Execute: func(ctx *WorkflowContext, payload *Payload, data *Metadata) (*StepResult, error) {
			if len(data.ReservedProducts) == 0 {
				return nil, errors.New("no reserved products found")
			}

			if err := s.activities.ConfirmProductsDeduction(ctx.Context(), payload.Order, data.ReservedProducts); err != nil {
				return nil, err
			}

			return &StepResult{Success: true}, nil
		},
		Compensate: func(ctx *WorkflowContext, payload *Payload, _ *Metadata) error {
			return s.activities.RestoreProducts(ctx.Context(), payload.Order)
		},
	})

	// Step 9: Send Order Confirmation Notifications
	executor.AddStep(&Step{
		Name:        constant.SendOrderConfirmedNotificationStep,
		Description: "Send order confirmation and receipt to customer after fulfillment created and order paid; includes invoice and tracking info",
		MaxRetries:  constant.SendOrderConfirmedNotificationStepMaxRetries,
		RetryDelay:  constant.SendOrderConfirmedNotificationStepRetryDelay,
		Timeout:     constant.SendOrderConfirmedNotificationStepTimeout,
		Idempotent:  true,
		Critical:    false,
		Execute: func(ctx *WorkflowContext, payload *Payload, data *Metadata) (*StepResult, error) {
			if data.TrackingNumber == nil {
				ctx.logger.Warn("No tracking number found for notification")

				return nil, errors.New("no tracking number found")
			}

			if len(data.ReservedProducts) == 0 {
				ctx.logger.Error("No reserved products found for notification")

				return nil, errors.New("no reserved products found")
			}

			if data.CustomerEmail == "" {
				ctx.logger.Error("No customer email found for notification")

				return nil, errors.New("no customer email found")
			}

			if err := s.activities.SendOrderConfirmedNotification(ctx.Context(), payload.Order, data.ReservedProducts, data.TrackingNumber, data.CustomerEmail); err != nil {
				// Non-critical step, log but don't fail the saga
				ctx.logger.Warnf("Failed to send notification: %v", err)
			}

			return &StepResult{Success: true}, nil
		},
		Compensate: nil, // No compensation for notifications
	})
}
