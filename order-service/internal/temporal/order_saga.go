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

// executeSagaSteps executes all saga steps in order to match saga implementation.
func executeSagaSteps(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	userAuth pkgdto.UserAuthInfo,
) error {
	steps := []stepHandler{
		{name: "ReserveProducts", handler: executeReserveProductsStep, critical: true},
		{name: "GetShippingCost", handler: executeGetShippingCostStep, critical: false},
		{name: "SetFinalPrices", handler: executeSetFinalPricesStep, critical: false},
		{name: "CreatePayment", handler: executeCreatePaymentStep, critical: true},
		{
			name:     "SendPaymentNotification",
			handler:  executeSendPaymentNotificationStep,
			critical: true,
		},
		{name: "WaitForPayment", handler: executeWaitForPaymentStep, critical: true},
		{name: "ProcessFulfillment", handler: executeProcessFulfillmentStep, critical: false},
		{name: "ConfirmDeduction", handler: executeConfirmDeductionStep, critical: true},
		{name: "SendConfirmation", handler: executeSendConfirmationStep, critical: false},
	}

	for _, step := range steps {
		if err := step.handler(ctx, order, state, userAuth); err != nil {
			if step.critical {
				return err
			}
			// Log non-critical errors but continue
			logger := workflow.GetLogger(ctx)
			logger.Warn("Non-critical step failed, continuing", "step", step.name, "error", err)
		}
	}

	return nil
}

// stepHandler represents a workflow step with its execution function and metadata.
//
// Fields:
//   - name: Human-readable name for logging and debugging
//   - handler: Function that executes the step logic
//   - critical: Whether step failure should trigger saga compensation
type stepHandler struct {
	name     string
	handler  stepExecutor
	critical bool
}

// stepExecutor defines the signature for step execution functions.
// All step executors receive the same parameters for consistency and access to needed data.
//
// Parameters:
//   - ctx: Temporal workflow context for activity execution
//   - order: The order being processed (may be modified during execution)
//   - state: Workflow state for tracking progress and storing intermediate results
//   - userAuth: User authentication information for service calls
//   - shipping: Shipping information for cost calculation and fulfillment
//
// Returns:
//   - error: Non-nil if the step fails and should be retried or compensated
type stepExecutor func(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	userAuth pkgdto.UserAuthInfo,
) error

// executeReserveProductsStep handles product reservation and calculation.
func executeReserveProductsStep(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	userAuth pkgdto.UserAuthInfo,
) error {
	logger := workflow.GetLogger(ctx)
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

	return nil
}

// executeGetShippingCostStep handles shipping cost calculation.
func executeGetShippingCostStep(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	userAuth pkgdto.UserAuthInfo,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing GetShippingCost", "orderID", order.ID)

	shippingRequest := dto.GetShippingCostRequest{
		Order:    order,
		UserAuth: &userAuth,
	}

	var shippingResult dto.GetShippingCostResponse
	if err := workflow.ExecuteActivity(ctx, string(constant.GetShippingCostStep), shippingRequest).Get(ctx, &shippingResult); err != nil {
		return err
	}

	state.ShippingCost = &shippingResult.ShippingCost
	state.CompletedSteps[constant.GetShippingCostStep] = true

	return nil
}

// executeSetFinalPricesStep handles final order price calculation.
func executeSetFinalPricesStep(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	_ pkgdto.UserAuthInfo,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing SetFinalOrderPrices", "orderID", order.ID)

	if state.ShippingCost == nil {
		return fmt.Errorf("shipping cost not available for order %s", order.ID)
	}

	setPricesInput := dto.SetFinalOrderPricesRequest{
		Order:        order,
		ShippingCost: *state.ShippingCost,
	}

	var setPricesResult dto.SetFinalOrderPricesResponse
	if err := workflow.ExecuteActivity(ctx, string(constant.SetFinalPricesStep), setPricesInput).Get(ctx, &setPricesResult); err != nil {
		return err
	}

	// Update the order with the latest data from database
	*order = *setPricesResult.UpdatedOrder
	state.CompletedSteps[constant.SetFinalPricesStep] = true

	return nil
}

// executeCreatePaymentStep handles payment creation.
func executeCreatePaymentStep(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	_ pkgdto.UserAuthInfo,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing CreatePayment", "orderID", order.ID)

	var paymentID uuid.UUID
	if err := workflow.ExecuteActivity(ctx, string(constant.CreatePaymentStep), order).Get(ctx, &paymentID); err != nil {
		return err
	}

	state.PaymentID = &paymentID
	state.CompletedSteps[constant.CreatePaymentStep] = true

	return nil
}

// executeSendPaymentNotificationStep handles payment required notification.
func executeSendPaymentNotificationStep(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	_ pkgdto.UserAuthInfo,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing SendPaymentRequiredNotification", "orderID", order.ID)

	if len(state.ReservedProducts) == 0 || state.CustomerEmail == "" {
		return fmt.Errorf("missing required data for payment notification: products=%d, email=%s",
			len(state.ReservedProducts), state.CustomerEmail)
	}

	paymentNotificationInput := dto.SendPaymentRequiredNotificationRequest{
		Order:            order,
		ReservedProducts: state.ReservedProducts,
		CustomerEmail:    state.CustomerEmail,
	}

	if err := workflow.ExecuteActivity(ctx, string(constant.SendPaymentRequiredNotificationStep), paymentNotificationInput).Get(ctx, nil); err != nil {
		return err
	}

	state.CompletedSteps[constant.SendPaymentRequiredNotificationStep] = true

	return nil
}

// executeWaitForPaymentStep handles payment confirmation with reminders.
func executeWaitForPaymentStep(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	_ pkgdto.UserAuthInfo,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Waiting for payment confirmation with reminders", "orderID", order.ID)

	var paymentConfirmationResult dto.WaitForPaymentConfirmationResponse

	paymentReceived := false

	// Setup payment confirmation with reminders
	if err := executePaymentConfirmationWithReminders(ctx, order, state, &paymentConfirmationResult, &paymentReceived); err != nil {
		return err
	}

	// Update payment ID from confirmation
	state.PaymentID = &paymentConfirmationResult.PaymentID
	state.CompletedSteps[constant.WaitForPaymentConfirmationStep] = true

	return nil
}

// executeProcessFulfillmentStep handles fulfillment processing.
func executeProcessFulfillmentStep(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	_ pkgdto.UserAuthInfo,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing ProcessFulfillment", "orderID", order.ID)

	var fulfillmentResult dto.ProcessFulfillmentResponse
	if err := workflow.ExecuteActivity(ctx, string(constant.ProcessFulfillmentStep), order).Get(ctx, &fulfillmentResult); err != nil {
		return err
	}

	state.ShippingID = &fulfillmentResult.ShippingID
	state.TrackingNumber = &fulfillmentResult.TrackingNumber
	state.CompletedSteps[constant.ProcessFulfillmentStep] = true

	return nil
}

// executeConfirmDeductionStep handles product deduction confirmation.
func executeConfirmDeductionStep(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	userAuth pkgdto.UserAuthInfo,
) error {
	logger := workflow.GetLogger(ctx)
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

	return nil
}

// executeSendConfirmationStep handles order confirmation notification.
func executeSendConfirmationStep(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	_ pkgdto.UserAuthInfo,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing SendOrderConfirmedNotification", "orderID", order.ID)

	if state.TrackingNumber == nil || state.CustomerEmail == "" {
		return fmt.Errorf("missing required data for order confirmation: tracking=%v, email=%s",
			state.TrackingNumber, state.CustomerEmail)
	}

	confirmationInput := dto.SendOrderConfirmedNotificationRequest{
		Order:          order,
		Products:       state.ReservedProducts,
		TrackingNumber: *state.TrackingNumber,
		CustomerEmail:  state.CustomerEmail,
	}

	if err := workflow.ExecuteActivity(ctx, string(constant.SendOrderConfirmedNotificationStep), confirmationInput).Get(ctx, nil); err != nil {
		return err
	}

	state.CompletedSteps[constant.SendOrderConfirmedNotificationStep] = true

	return nil
}

// executePaymentConfirmationWithReminders handles payment confirmation with reminder logic.
func executePaymentConfirmationWithReminders(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	paymentConfirmationResult *dto.WaitForPaymentConfirmationResponse,
	paymentReceived *bool,
) error {
	// Create timer channels for reminders
	firstReminderTimer := workflow.NewTimer(ctx, constant.FirstPaymentReminderDelay)
	secondReminderTimer := workflow.NewTimer(ctx, constant.SecondPaymentReminderDelay)
	expireOrderTimer := workflow.NewTimer(ctx, constant.ExpireOrderReminderDelay)

	selector := workflow.NewSelector(ctx)

	// Add timer callbacks for reminders
	selector.AddFuture(firstReminderTimer, func(_ workflow.Future) {
		if !*paymentReceived {
			sendPaymentReminder(
				ctx,
				order,
				state,
				constant.FirstReminderSequence,
				constant.FirstPaymentReminderEmailSubject,
			)
		}
	})

	selector.AddFuture(secondReminderTimer, func(_ workflow.Future) {
		if !*paymentReceived {
			sendPaymentReminder(
				ctx,
				order,
				state,
				constant.SecondReminderSequence,
				constant.SecondPaymentReminderEmailSubject,
			)
		}
	})

	selector.AddFuture(expireOrderTimer, func(_ workflow.Future) {
		if !*paymentReceived {
			logger := workflow.GetLogger(ctx)
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
		if err := f.Get(ctx, paymentConfirmationResult); err == nil {
			*paymentReceived = true

			logger := workflow.GetLogger(ctx)
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
	for !*paymentReceived {
		selector.Select(ctx)

		// Check if order expired
		if expireOrderTimer.IsReady() {
			return fmt.Errorf("payment timeout reached for order %s", order.ID)
		}

		// If payment received, break the loop
		if *paymentReceived {
			break
		}
	}

	return nil
}

// sendPaymentReminder sends a payment reminder notification.
func sendPaymentReminder(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	reminderSequence int,
	subject string,
) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Sending payment reminder", "orderID", order.ID, "sequence", reminderSequence)

	reminderRequest := dto.SendPaymentReminderNotificationRequest{
		Order:            order,
		ReservedProducts: state.ReservedProducts,
		CustomerEmail:    state.CustomerEmail,
		ReminderSequence: reminderSequence,
		Subject:          subject,
	}

	err := workflow.ExecuteActivity(ctx, string(constant.SendPaymentReminderNotificationStep), reminderRequest).
		Get(ctx, nil)
	if err != nil {
		logger.Warn(
			"Failed to send payment reminder",
			"orderID",
			order.ID,
			"sequence",
			reminderSequence,
			"error",
			err,
		)
	}
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

	// Define compensation steps in reverse order
	compensationSteps := []compensationStep{
		{
			step:     constant.ProcessFulfillmentStep,
			name:     "CancelShipping",
			handler:  executeCancelShippingCompensation,
			critical: false,
		},
		{
			step:     constant.ConfirmProductsDeductionStep,
			name:     "RestoreProducts",
			handler:  executeRestoreProductsCompensation,
			critical: false,
		},
		{
			step:     constant.CreatePaymentStep,
			name:     "RefundPayment",
			handler:  executeRefundPaymentCompensation,
			critical: true, // Payment refund is critical
		},
		{
			step:     constant.ReserveProductsStep,
			name:     "ReleaseProducts",
			handler:  executeReleaseProductsCompensation,
			critical: false,
		},
	}

	// Configure compensation activity options
	compensationOptions := workflow.ActivityOptions{
		StartToCloseTimeout: constant.TemporalCompensationWorkflowTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: constant.TemporalBackoffCoefficient,
			MaximumInterval:    constant.TemporalMaxInterval,
			MaximumAttempts:    constant.TemporalMaxAttempts,
		},
	}

	compensationCtx := workflow.WithActivityOptions(ctx, compensationOptions)

	var (
		compensationErrors []string
		criticalError      error
	)

	// Execute compensation steps
	for _, compStep := range compensationSteps {
		if !state.CompletedSteps[compStep.step] {
			continue // Skip compensation if step wasn't completed
		}

		logger.Info("Executing compensation", "step", compStep.name, "orderID", order.ID)

		if err := compStep.handler(compensationCtx, order, state, userAuth); err != nil {
			logger.Error(
				"Compensation failed",
				"step",
				compStep.name,
				"error",
				err,
				"orderID",
				order.ID,
			)
			compensationErrors = append(
				compensationErrors,
				fmt.Sprintf("%s: %v", compStep.name, err),
			)

			if compStep.critical {
				criticalError = err
			}
		} else {
			logger.Info("Compensation completed", "step", compStep.name, "orderID", order.ID)
		}
	}

	// Handle compensation results
	if len(compensationErrors) > 0 {
		logger.Warn("Compensation completed with errors",
			"orderID", order.ID,
			"errorCount", len(compensationErrors),
			"errors", compensationErrors)

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

// compensationStep represents a compensation step with metadata.
//
// Fields:
//   - step: The original workflow step that this compensation corresponds to
//   - name: Human-readable name for logging and debugging
//   - handler: Function that executes the compensation logic
//   - critical: Whether compensation failure should fail the entire compensation process
type compensationStep struct {
	step     constant.WorkflowStep
	name     string
	handler  compensationHandler
	critical bool
}

// compensationHandler defines the signature for compensation functions.
// Compensation functions should be idempotent and handle cases where the original step
// was only partially completed or where multiple compensation attempts are made.
//
// Parameters:
//   - ctx: Temporal workflow context for activity execution
//   - order: The order being compensated
//   - state: Workflow state containing data from completed steps
//   - userAuth: User authentication information for service calls
//
// Returns:
//   - error: Non-nil if compensation fails and should be retried
type compensationHandler func(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	userAuth pkgdto.UserAuthInfo,
) error

// executeCancelShippingCompensation handles shipping cancellation.
func executeCancelShippingCompensation(
	ctx workflow.Context,
	_ *entity.Order,
	state *dto.TemporalWorkflowState,
	_ pkgdto.UserAuthInfo,
) error {
	if state.ShippingID == nil {
		return nil // No shipping to cancel
	}

	return workflow.ExecuteActivity(ctx, string(constant.CancelShippingStep), *state.ShippingID).
		Get(ctx, nil)
}

// executeRestoreProductsCompensation handles product restoration.
func executeRestoreProductsCompensation(
	ctx workflow.Context,
	order *entity.Order,
	_ *dto.TemporalWorkflowState,
	userAuth pkgdto.UserAuthInfo,
) error {
	restoreReq := dto.RestoreProductsRequest{
		Order:    order,
		UserAuth: userAuth,
	}

	return workflow.ExecuteActivity(ctx, string(constant.RestoreProductsStep), restoreReq).
		Get(ctx, nil)
}

// executeRefundPaymentCompensation handles payment refund.
func executeRefundPaymentCompensation(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
	_ pkgdto.UserAuthInfo,
) error {
	if state.PaymentID == nil {
		return nil // No payment to refund
	}

	refundInput := dto.RefundPaymentGatewayRequest{
		Order:     order,
		PaymentID: *state.PaymentID,
	}

	return workflow.ExecuteActivity(ctx, string(constant.RefundPaymentStep), refundInput).
		Get(ctx, nil)
}

// executeReleaseProductsCompensation handles product release.
func executeReleaseProductsCompensation(
	ctx workflow.Context,
	order *entity.Order,
	_ *dto.TemporalWorkflowState,
	userAuth pkgdto.UserAuthInfo,
) error {
	releaseReq := dto.ReleaseProductsRequest{
		Order:    order,
		UserAuth: userAuth,
	}

	return workflow.ExecuteActivity(ctx, string(constant.ReleaseProductsStep), releaseReq).
		Get(ctx, nil)
}
