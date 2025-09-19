package temporal

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

//nolint:funlen,gocyclo,cyclop // executeSagaSteps executes all saga steps in order to match saga implementation.
func executeSagaSteps(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	userAuth pkgdto.UserAuthInfo,
	shipping *dto.Shipping,
) error {
	logger := workflow.GetLogger(ctx)

	// Step 1: Reserve Products and Calculate
	logger.Info("Executing ReserveProducts", "orderID", order.ID)

	reserveRequest := dto.ReserveProductsRequest{
		Order:    order,
		UserAuth: userAuth,
	}

	var reserveResult dto.ReserveProductsResponse
	if err := workflow.ExecuteActivity(ctx, string(constant.ReserveProductsStep), reserveRequest).Get(ctx, &reserveResult); err != nil {
		return temporal.NewNonRetryableApplicationError(
			"ReserveProducts failed",
			"ReserveProductsError",
			err,
		)
	}

	// Update order with calculated items
	order.Items = reserveResult.CalculatedOrder.Items
	state.ReservedProducts = reserveResult.ReservedProducts
	state.CustomerEmail = reserveResult.CustomerEmail
	state.CompletedSteps[constant.ReserveProductsStep] = true

	// Step 2: Get Shipping Cost
	logger.Info("Executing GetShippingCost", "orderID", order.ID)

	shippingRequest := dto.GetShippingCostRequest{
		Order:    order,
		Shipping: shipping,
		UserAuth: &userAuth,
	}

	var shippingResult dto.GetShippingCostResponse
	if err := workflow.ExecuteActivity(ctx, string(constant.GetShippingCostStep), shippingRequest).Get(ctx, &shippingResult); err != nil {
		// Non-critical step, log but continue
		logger.Warn(
			"GetShippingCost failed, but saga will continue",
			"error",
			err,
			"orderID",
			order.ID,
		)
	} else {
		state.ShippingCost = &shippingResult.ShippingCost
		state.CompletedSteps[constant.GetShippingCostStep] = true
	}

	// Step 3: Set Final Order Prices
	logger.Info("Executing SetFinalOrderPrices", "orderID", order.ID)

	if state.ShippingCost != nil {
		setPricesInput := dto.SetFinalOrderPricesRequest{
			Order:        order,
			ShippingCost: *state.ShippingCost,
		}

		var setPricesResult dto.SetFinalOrderPricesResponse

		if err := workflow.ExecuteActivity(ctx, string(constant.SetFinalPricesStep), setPricesInput).Get(ctx, &setPricesResult); err != nil {
			// Non-critical step, log but continue
			logger.Warn(
				"SetFinalOrderPrices failed, but saga will continue",
				"error",
				err,
				"orderID",
				order.ID,
			)
		} else {
			// Update the order with the latest data from database
			order = setPricesResult.UpdatedOrder
			state.CompletedSteps[constant.SetFinalPricesStep] = true
		}
	}

	// Step 4: Create Payment
	logger.Info("Executing CreatePayment", "orderID", order.ID)

	var paymentID uuid.UUID
	if err := workflow.ExecuteActivity(ctx, string(constant.CreatePaymentStep), order).Get(ctx, &paymentID); err != nil {
		return err
	}

	state.PaymentID = &paymentID
	state.CompletedSteps[constant.CreatePaymentStep] = true

	// Step 5: Send Payment Required Notification
	logger.Info("Executing SendPaymentRequiredNotification", "orderID", order.ID)

	if len(state.ReservedProducts) > 0 && state.CustomerEmail != "" {
		paymentNotificationInput := dto.SendPaymentRequiredNotificationRequest{
			Order:            order,
			ReservedProducts: state.ReservedProducts,
			CustomerEmail:    state.CustomerEmail,
		}
		if err := workflow.ExecuteActivity(ctx, string(constant.SendPaymentRequiredNotificationStep), paymentNotificationInput).Get(ctx, nil); err != nil {
			return err
		}

		state.CompletedSteps[constant.SendPaymentRequiredNotificationStep] = true
	}

	// Step 6: Wait for Payment Confirmation with embedded reminders
	logger.Info("Waiting for payment confirmation with reminders", "orderID", order.ID)

	var paymentConfirmationResult dto.WaitForPaymentConfirmationResponse

	paymentReceived := false

	// Create timer channels for reminders
	firstReminderTimer := workflow.NewTimer(ctx, constant.FirstPaymentReminderDelay)
	secondReminderTimer := workflow.NewTimer(ctx, constant.SecondPaymentReminderDelay)
	expireOrderTimer := workflow.NewTimer(ctx, constant.ExpireOrderReminderDelay)

	selector := workflow.NewSelector(ctx)

	// Add timer callbacks for reminders
	selector.AddFuture(firstReminderTimer, func(_ workflow.Future) {
		if !paymentReceived {
			logger.Info("Sending first payment reminder", "orderID", order.ID)
			reminderRequest := dto.SendPaymentReminderNotificationRequest{
				Order:            order,
				ReservedProducts: state.ReservedProducts,
				CustomerEmail:    state.CustomerEmail,
				ReminderSequence: constant.FirstReminderSequence,
				Subject:          constant.FirstPaymentReminderEmailSubject,
			}

			err := workflow.ExecuteActivity(ctx, string(constant.SendPaymentReminderNotificationStep), reminderRequest).
				Get(ctx, nil)
			if err != nil {
				logger.Warn(
					"Failed to send first payment reminder",
					"orderID",
					order.ID,
					"error",
					err,
				)
			}
		}
	})

	selector.AddFuture(secondReminderTimer, func(_ workflow.Future) {
		if !paymentReceived {
			logger.Info("Sending second payment reminder", "orderID", order.ID)
			reminderRequest := dto.SendPaymentReminderNotificationRequest{
				Order:            order,
				ReservedProducts: state.ReservedProducts,
				CustomerEmail:    state.CustomerEmail,
				ReminderSequence: constant.SecondReminderSequence,
				Subject:          constant.SecondPaymentReminderEmailSubject,
			}

			err := workflow.ExecuteActivity(ctx, string(constant.SendPaymentReminderNotificationStep), reminderRequest).
				Get(ctx, nil)
			if err != nil {
				logger.Warn(
					"Failed to send second payment reminder",
					"orderID",
					order.ID,
					"error",
					err,
				)
			}
		}
	})

	selector.AddFuture(expireOrderTimer, func(_ workflow.Future) {
		if !paymentReceived {
			logger.Info("Payment timeout reached, failing saga", "orderID", order.ID)
		}
	})

	// Start payment wait activity with no retries
	paymentActivityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: constant.WaitForPaymentConfirmationStepTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	paymentCtx := workflow.WithActivityOptions(ctx, paymentActivityOptions)

	paymentFuture := workflow.ExecuteActivity(
		paymentCtx,
		string(constant.WaitForPaymentConfirmationStep),
		dto.WaitForPaymentConfirmationRequest{Order: order},
	)

	// Add payment confirmation callback
	selector.AddFuture(paymentFuture, func(f workflow.Future) {
		if err := f.Get(ctx, &paymentConfirmationResult); err == nil {
			paymentReceived = true

			logger.Info(
				"Payment confirmation received",
				"orderID",
				order.ID,
				"paymentID",
				paymentConfirmationResult.PaymentID,
			)
		}
	})

	// Wait for either payment or final timeout
	for !paymentReceived {
		selector.Select(ctx)

		// Check if order expired
		if expireOrderTimer.IsReady() {
			return fmt.Errorf("payment timeout reached for order %s", order.ID)
		}

		// If payment received, break the loop
		if paymentReceived {
			break
		}
	}

	// Update payment ID from confirmation
	state.PaymentID = &paymentConfirmationResult.PaymentID
	state.CompletedSteps[constant.WaitForPaymentConfirmationStep] = true

	// Step 7: Process Fulfillment
	logger.Info("Executing ProcessFulfillment", "orderID", order.ID)

	var fulfillmentResult dto.ProcessFulfillmentResponse
	if err := workflow.ExecuteActivity(ctx, string(constant.ProcessFulfillmentStep), order, shipping).Get(ctx, &fulfillmentResult); err != nil {
		// Non-critical step, log but continue
		logger.Warn(
			"ProcessFulfillment failed, but saga will continue",
			"error",
			err,
			"orderID",
			order.ID,
		)
	} else {
		state.ShippingID = &fulfillmentResult.ShippingID
		state.TrackingNumber = &fulfillmentResult.TrackingNumber
		state.CompletedSteps[constant.ProcessFulfillmentStep] = true
	}

	// Step 8: Confirm Products Deduction
	logger.Info("Executing ConfirmProductsDeduction", "orderID", order.ID)

	confirmDeductionInput := dto.ConfirmProductsDeductionRequest{
		Order:            order,
		ReservedProducts: state.ReservedProducts,
		UserAuth:         userAuth,
	}
	if err := workflow.ExecuteActivity(ctx, string(constant.ConfirmProductsDeductionStep), confirmDeductionInput).Get(ctx, nil); err != nil {
		return err
	}

	state.CompletedSteps[constant.ConfirmProductsDeductionStep] = true

	// Step 9: Send Order Confirmation
	logger.Info("Executing SendOrderConfirmedNotification", "orderID", order.ID)

	if state.TrackingNumber != nil && state.CustomerEmail != "" {
		confirmationInput := dto.SendOrderConfirmedNotificationRequest{
			Order:          order,
			Products:       state.ReservedProducts,
			TrackingNumber: *state.TrackingNumber,
			CustomerEmail:  state.CustomerEmail,
		}
		if err := workflow.ExecuteActivity(ctx, string(constant.SendOrderConfirmedNotificationStep), confirmationInput).Get(ctx, nil); err != nil {
			// This is not critical, log but don't fail the saga
			logger.Warn(
				"SendOrderConfirmedNotification failed, but saga will continue",
				"error",
				err,
				"orderID",
				order.ID,
			)
		} else {
			state.CompletedSteps[constant.SendOrderConfirmedNotificationStep] = true
		}
	}

	return nil
}

// executeCompensation executes compensation activities in reverse order to match saga implementation.
func executeCompensation(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	userAuth pkgdto.UserAuthInfo,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting compensation", "orderID", order.ID)

	var (
		compensationErrors []string
		criticalError      error
	)

	// Configure compensation activity options with shorter timeout
	compensationOptions := workflow.ActivityOptions{
		StartToCloseTimeout: constant.TemporalWorkflowTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: constant.TemporalBackoffCoefficient,
			MaximumInterval:    constant.TemporalMaxInterval,
			MaximumAttempts:    constant.TemporalMaxAttempts,
		},
	}

	compensationCtx := workflow.WithActivityOptions(ctx, compensationOptions)

	// Compensation in reverse order of saga execution

	// Cancel Shipping (if shipping was created) - Step 7 compensation
	if state.CompletedSteps[constant.ProcessFulfillmentStep] && state.ShippingID != nil {
		logger.Info(
			"Compensating ProcessFulfillment (Step 7)",
			"orderID",
			order.ID,
			"shippingID",
			*state.ShippingID,
		)

		if err := workflow.ExecuteActivity(compensationCtx, string(constant.CancelShippingStep), *state.ShippingID).Get(compensationCtx, nil); err != nil {
			logger.Error("CancelShipping compensation failed", "error", err, "orderID", order.ID)
			compensationErrors = append(compensationErrors, "CancelShipping: "+err.Error())
		}
	}

	// Restore Products (if products were deducted) - Step 8 compensation
	if state.CompletedSteps[constant.ConfirmProductsDeductionStep] {
		logger.Info("Compensating ConfirmProductsDeduction (Step 8)", "orderID", order.ID)

		restoreReq := dto.RestoreProductsRequest{
			Order:    order,
			UserAuth: userAuth,
		}
		if err := workflow.ExecuteActivity(compensationCtx, string(constant.RestoreProductsStep), restoreReq).Get(compensationCtx, nil); err != nil {
			logger.Error("RestoreProducts compensation failed", "error", err, "orderID", order.ID)
			compensationErrors = append(compensationErrors, "RestoreProducts: "+err.Error())
		}
	}

	// Refund Payment (if payment was processed) - Step 4 & 6 compensation
	if (state.CompletedSteps[constant.CreatePaymentStep] || state.CompletedSteps[constant.WaitForPaymentConfirmationStep]) &&
		state.PaymentID != nil {
		logger.Info(
			"Compensating CreatePayment/WaitForPaymentConfirmation (Step 4/6)",
			"orderID",
			order.ID,
			"paymentID",
			*state.PaymentID,
		)

		refundInput := dto.RefundPaymentGatewayRequest{
			Order:     order,
			PaymentID: *state.PaymentID,
		}
		if err := workflow.ExecuteActivity(compensationCtx, string(constant.RefundPaymentStep), refundInput).Get(compensationCtx, nil); err != nil {
			logger.Error("RefundPayment compensation failed", "error", err, "orderID", order.ID)
			compensationErrors = append(compensationErrors, "RefundPayment: "+err.Error())
			// Payment refund failure is critical - we need to track this for manual intervention
			criticalError = err
		}
	}

	// Release Products (if products were reserved) - Step 1 compensation
	if state.CompletedSteps[constant.ReserveProductsStep] {
		logger.Info("Compensating ReserveProducts (Step 1)", "orderID", order.ID)

		releaseReq := dto.ReleaseProductsRequest{
			Order:    order,
			UserAuth: userAuth,
		}
		if err := workflow.ExecuteActivity(compensationCtx, string(constant.ReleaseProductsStep), releaseReq).Get(compensationCtx, nil); err != nil {
			logger.Error("ReleaseProducts compensation failed", "error", err, "orderID", order.ID)
			compensationErrors = append(compensationErrors, "ReleaseProducts: "+err.Error())
		}
	}

	if len(compensationErrors) > 0 {
		logger.Warn("Compensation completed with errors",
			"orderID", order.ID,
			"errorCount", len(compensationErrors),
			"errors", compensationErrors)

		// Return critical error if payment refund failed, otherwise return first error
		if criticalError != nil {
			return fmt.Errorf("critical compensation failure: %w", criticalError)
		}

		return fmt.Errorf(
			"compensation completed with %d errors: %v",
			len(compensationErrors),
			compensationErrors,
		)
	}

	logger.Info("Compensation completed successfully", "orderID", order.ID)

	return nil
}
