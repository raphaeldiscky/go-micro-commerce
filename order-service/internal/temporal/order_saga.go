package temporal

import (
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

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
	logger := workflow.GetLogger(ctx)

	// Step 1: Reserve Products and Calculate
	logger.Info("Executing ReserveProductsAndCalculate", "orderID", order.ID)

	// Create activity request with user auth info
	reserveRequest := dto.ReserveProductsAndCalculateRequest{
		Order:    order,
		UserAuth: userAuth,
	}

	var reserveResult dto.ReserveProductsAndCalculateResponse
	if err := workflow.ExecuteActivity(ctx, ReserveProductsAndCalculate, reserveRequest).Get(ctx, &reserveResult); err != nil {
		return temporal.NewNonRetryableApplicationError(
			"ReserveProductsAndCalculate failed",
			"ReserveProductsAndCalculateError",
			err,
		)
	}

	// Update order with calculated items
	order.Items = reserveResult.CalculatedOrder.Items
	state.ReservedProducts = reserveResult.ReservedProducts
	state.CustomerEmail = reserveResult.CustomerEmail
	state.CompletedSteps["ReserveProductsAndCalculate"] = true

	// Step 2: Process Fulfillment
	logger.Info("Executing ProcessFulfillment", "orderID", order.ID)

	var fulfillmentResult dto.ProcessFulfillmentResponse
	if err := workflow.ExecuteActivity(ctx, ProcessFulfillment, order).Get(ctx, &fulfillmentResult); err != nil {
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
		state.ShippingCost = &fulfillmentResult.ShippingCost
		state.CompletedSteps["ProcessFulfillment"] = true
	}

	// Step 3: Set Final Order Prices
	logger.Info("Executing SetFinalOrderPrices", "orderID", order.ID)

	if state.ShippingCost != nil {
		setPricesInput := dto.SetFinalOrderPricesRequest{
			Order:        order,
			ShippingCost: *state.ShippingCost,
		}

		var setPricesResult dto.SetFinalOrderPricesResponse

		if err := workflow.ExecuteActivity(ctx, SetFinalOrderPrices, setPricesInput).Get(ctx, &setPricesResult); err != nil {
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
			state.CompletedSteps["SetFinalOrderPrices"] = true
		}
	}

	// Step 4: Process Payment
	logger.Info("Executing ProcessPayment", "orderID", order.ID)

	var paymentID uuid.UUID
	if err := workflow.ExecuteActivity(ctx, ProcessPayment, order).Get(ctx, &paymentID); err != nil {
		return err
	}

	state.PaymentID = &paymentID
	state.CompletedSteps["ProcessPayment"] = true

	// Step 5: Confirm Products Deduction
	logger.Info("Executing ConfirmProductsDeduction", "orderID", order.ID)

	confirmDeductionInput := dto.ConfirmProductsDeductionRequest{
		Order:            order,
		ReservedProducts: state.ReservedProducts,
		UserAuth:         userAuth,
	}
	if err := workflow.ExecuteActivity(ctx, ConfirmProductsDeduction, confirmDeductionInput).Get(ctx, nil); err != nil {
		return err
	}

	state.CompletedSteps["ConfirmProductsDeduction"] = true

	// Step 6: Send Order Confirmation
	logger.Info("Executing SendOrderConfirmedNotification", "orderID", order.ID)

	if state.TrackingNumber != nil && state.CustomerEmail != "" {
		confirmationInput := dto.SendOrderConfirmedNotificationRequest{
			Order:          order,
			Products:       state.ReservedProducts,
			TrackingNumber: *state.TrackingNumber,
			CustomerEmail:  state.CustomerEmail,
		}
		if err := workflow.ExecuteActivity(ctx, SendOrderConfirmedNotification, confirmationInput).Get(ctx, nil); err != nil {
			// This is not critical, log but don't fail the saga
			logger.Warn(
				"SendOrderConfirmedNotification failed, but saga will continue",
				"error",
				err,
				"orderID",
				order.ID,
			)
		} else {
			state.CompletedSteps["SendOrderConfirmedNotification"] = true
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

	// Cancel Shipping (if shipping was created)
	if state.CompletedSteps["ProcessFulfillment"] && state.ShippingID != nil {
		logger.Info(
			"Compensating ProcessFulfillment",
			"orderID",
			order.ID,
			"shippingID",
			*state.ShippingID,
		)

		if err := workflow.ExecuteActivity(compensationCtx, CancelShipping, *state.ShippingID).Get(compensationCtx, nil); err != nil {
			logger.Error("CancelShipping compensation failed", "error", err, "orderID", order.ID)
		}
	}

	// Restore Products (if products were deducted)
	if state.CompletedSteps["ConfirmProductsDeduction"] {
		logger.Info("Compensating ConfirmProductsDeduction", "orderID", order.ID)

		restoreReq := dto.RestoreProductsRequest{
			Order:    order,
			UserAuth: userAuth,
		}
		if err := workflow.ExecuteActivity(compensationCtx, RestoreProducts, restoreReq).Get(compensationCtx, nil); err != nil {
			logger.Error("RestoreProducts compensation failed", "error", err, "orderID", order.ID)
		}
	}

	// Refund Payment (if payment was processed)
	if state.CompletedSteps["ProcessPayment"] && state.PaymentID != nil {
		logger.Info(
			"Compensating ProcessPayment",
			"orderID",
			order.ID,
			"paymentID",
			*state.PaymentID,
		)

		refundInput := dto.RefundPaymentGatewayRequest{
			Order:     order,
			PaymentID: *state.PaymentID,
		}
		if err := workflow.ExecuteActivity(compensationCtx, RefundPayment, refundInput).Get(compensationCtx, nil); err != nil {
			logger.Error("RefundPayment compensation failed", "error", err, "orderID", order.ID)
		}
	}

	// Release Products (if products were reserved)
	if state.CompletedSteps["ReserveProductsAndCalculate"] {
		logger.Info("Compensating ReserveProductsAndCalculate", "orderID", order.ID)

		releaseReq := dto.ReleaseProductsRequest{
			Order:    order,
			UserAuth: userAuth,
		}
		if err := workflow.ExecuteActivity(compensationCtx, ReleaseProducts, releaseReq).Get(compensationCtx, nil); err != nil {
			logger.Error("ReleaseProducts compensation failed", "error", err, "orderID", order.ID)
		}
	}

	logger.Info("Compensation completed", "orderID", order.ID)
}
