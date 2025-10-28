// Package saga provides PostPaymentSaga workflow implementation.
package saga

import (
	"errors"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// PostPaymentSaga implements the post-payment orchestration workflow.
// This saga handles Steps 7-9 after payment succeeds:
// - Step 1 (7): Process Fulfillment
// - Step 2 (8): Confirm Products Deduction
// - Step 3 (9): Send Order Confirmed Notification.
type PostPaymentSaga struct {
	activities OrderActivities
}

// NewPostPaymentSaga creates a new post-payment saga.
func NewPostPaymentSaga(activities OrderActivities) *PostPaymentSaga {
	return &PostPaymentSaga{
		activities: activities,
	}
}

// ConfigureSteps configures all steps for the post-payment saga.
func (s *PostPaymentSaga) ConfigureSteps(executor *Executor) {
	// Step 1: Process Fulfillment (Step 7 in original saga)
	executor.AddStep(&Step{
		Name:        constant.ProcessFulfillmentStep,
		Description: "Process fulfillment for the paid order; get fulfillmentID, shippingCost and trackingNumber",
		MaxRetries:  constant.PostPaymentFulfillmentMaxRetries,
		RetryDelay:  constant.PostPaymentFulfillmentRetryDelay,
		Timeout:     constant.ProcessFulfillmentStepTimeout,
		Idempotent:  true,
		Critical:    true, // Critical for post-payment saga with retry
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
					ReservedProducts: data.ReservedProducts,
					CustomerEmail:    data.CustomerEmail,
					UserAuth:         data.UserAuth,
					PaymentID:        data.PaymentID,
					FulfillmentID:    &fulfillmentID,
					TrackingNumber:   &trackingNumber,
					ShippingCost:     &shippingCost,
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

	// Step 2: Confirm Products Deduction (Step 8 in original saga)
	executor.AddStep(&Step{
		Name:        constant.ConfirmProductsDeductionStep,
		Description: "Permanently deduct products from inventory after successful payment",
		MaxRetries:  constant.PostPaymentProductsMaxRetries,
		RetryDelay:  constant.PostPaymentProductsRetryDelay,
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

			return &StepResult{
				Success: true,
				Data:    data,
			}, nil
		},
		Compensate: func(ctx *WorkflowContext, payload *Payload, _ *Metadata) error {
			return s.activities.RestoreProducts(ctx.Context(), payload.Order)
		},
	})

	// Step 3: Send Order Confirmed Notification (Step 9 in original saga)
	executor.AddStep(&Step{
		Name:        constant.SendOrderConfirmedNotificationStep,
		Description: "Send order confirmation and receipt to customer after fulfillment created and order paid; includes invoice and tracking info",
		MaxRetries:  constant.PostPaymentNotificationMaxRetries,
		RetryDelay:  constant.PostPaymentNotificationRetryDelay,
		Timeout:     constant.SendOrderConfirmedNotificationStepTimeout,
		Idempotent:  true,
		Critical:    false, // Notification failure shouldn't fail entire saga
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
				ctx.logger.Warnf("Failed to send notification: %v", err)
				// Don't fail saga if notification fails - order is already complete
			}

			return &StepResult{Success: true}, nil
		},
		Compensate: nil, // No compensation for notifications
	})
}
