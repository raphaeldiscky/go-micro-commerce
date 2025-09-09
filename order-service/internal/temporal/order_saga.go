package temporal

import (
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

//nolint:funlen // executeSagaSteps executes all saga steps in order to match saga implementation.
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
	if err := workflow.ExecuteActivity(ctx, constant.ReserveProductsStep, reserveRequest).Get(ctx, &reserveResult); err != nil {
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
	if err := workflow.ExecuteActivity(ctx, constant.GetShippingCostStep, shippingRequest).Get(ctx, &shippingResult); err != nil {
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

		if err := workflow.ExecuteActivity(ctx, constant.SetFinalPricesStep, setPricesInput).Get(ctx, &setPricesResult); err != nil {
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
	if err := workflow.ExecuteActivity(ctx, constant.CreatePaymentStep, order).Get(ctx, &paymentID); err != nil {
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
		if err := workflow.ExecuteActivity(ctx, constant.SendPaymentRequiredNotificationStep, paymentNotificationInput).Get(ctx, nil); err != nil {
			return err
		}

		state.CompletedSteps[constant.SendPaymentRequiredNotificationStep] = true
	}

	// Step 6: Wait for Payment Confirmation
	logger.Info("Executing WaitForPaymentConfirmation", "orderID", order.ID)

	waitPaymentRequest := dto.WaitForPaymentConfirmationRequest{
		Order: order,
	}

	var paymentConfirmationResult dto.WaitForPaymentConfirmationResponse
	if err := workflow.ExecuteActivity(ctx, constant.WaitForPaymentConfirmationStep, waitPaymentRequest).Get(ctx, &paymentConfirmationResult); err != nil {
		return err
	}

	// Update payment ID from confirmation
	state.PaymentID = &paymentConfirmationResult.PaymentID
	state.CompletedSteps[constant.WaitForPaymentConfirmationStep] = true

	// Step 7: Process Fulfillment
	logger.Info("Executing ProcessFulfillment", "orderID", order.ID)

	var fulfillmentResult dto.ProcessFulfillmentResponse
	if err := workflow.ExecuteActivity(ctx, constant.ProcessFulfillmentStep, order, shipping).Get(ctx, &fulfillmentResult); err != nil {
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
	if err := workflow.ExecuteActivity(ctx, constant.ConfirmProductsDeductionStep, confirmDeductionInput).Get(ctx, nil); err != nil {
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
		if err := workflow.ExecuteActivity(ctx, constant.SendOrderConfirmedNotificationStep, confirmationInput).Get(ctx, nil); err != nil {
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
) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting compensation", "orderID", order.ID)

	// Configure compensation activity options with shorter timeout
	compensationOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1.5,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    2,
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

		if err := workflow.ExecuteActivity(compensationCtx, constant.CancelShippingStep, *state.ShippingID).Get(compensationCtx, nil); err != nil {
			logger.Error("CancelShipping compensation failed", "error", err, "orderID", order.ID)
		}
	}

	// Restore Products (if products were deducted) - Step 8 compensation
	if state.CompletedSteps[constant.ConfirmProductsDeductionStep] {
		logger.Info("Compensating ConfirmProductsDeduction (Step 8)", "orderID", order.ID)

		restoreReq := dto.RestoreProductsRequest{
			Order:    order,
			UserAuth: userAuth,
		}
		if err := workflow.ExecuteActivity(compensationCtx, constant.RestoreProductsStep, restoreReq).Get(compensationCtx, nil); err != nil {
			logger.Error("RestoreProducts compensation failed", "error", err, "orderID", order.ID)
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
		if err := workflow.ExecuteActivity(compensationCtx, constant.RefundPaymentStep, refundInput).Get(compensationCtx, nil); err != nil {
			logger.Error("RefundPayment compensation failed", "error", err, "orderID", order.ID)
		}
	}

	// Release Products (if products were reserved) - Step 1 compensation
	if state.CompletedSteps[constant.ReserveProductsStep] {
		logger.Info("Compensating ReserveProducts (Step 1)", "orderID", order.ID)

		releaseReq := dto.ReleaseProductsRequest{
			Order:    order,
			UserAuth: userAuth,
		}
		if err := workflow.ExecuteActivity(compensationCtx, constant.ReleaseProductsStep, releaseReq).Get(compensationCtx, nil); err != nil {
			logger.Error("ReleaseProducts compensation failed", "error", err, "orderID", order.ID)
		}
	}

	logger.Info("Compensation completed", "orderID", order.ID)
}
