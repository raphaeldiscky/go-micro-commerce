package temporal

import (
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// executeSagaSteps executes all saga steps in order.
func executeSagaSteps(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
) error {
	logger := workflow.GetLogger(ctx)

	// Step 1: Validate Products
	logger.Info("Executing ValidateProducts", "orderID", order.ID)

	if err := workflow.ExecuteActivity(ctx, ValidateProducts, order).Get(ctx, nil); err != nil {
		return temporal.NewNonRetryableApplicationError(
			"ValidateProducts failed",
			"ValidateProductsError",
			err,
		)
	}

	state.CompletedSteps["ValidateProducts"] = true

	// Step 2: Reserve Products
	logger.Info("Executing ReserveProducts", "orderID", order.ID)

	var reservedProducts []entity.Product
	if err := workflow.ExecuteActivity(ctx, ReserveProducts, order).Get(ctx, &reservedProducts); err != nil {
		return err
	}

	state.ReservedProducts = reservedProducts
	state.CompletedSteps["ReserveProducts"] = true

	// Step 3: Calculate Pricing
	logger.Info("Executing CalculatePricing", "orderID", order.ID)

	var pricing entity.OrderPricing
	if err := workflow.ExecuteActivity(ctx, CalculatePricing, order).Get(ctx, &pricing); err != nil {
		return err
	}

	state.Pricing = &pricing
	state.CompletedSteps["CalculatePricing"] = true

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

	if err := workflow.ExecuteActivity(ctx, ConfirmProductsDeduction, order).Get(ctx, nil); err != nil {
		return err
	}

	state.CompletedSteps["ConfirmProductsDeduction"] = true

	// Step 6: Create Shipping
	logger.Info("Executing CreateShipping", "orderID", order.ID)

	var shippingResult dto.CreateShippingResponse
	if err := workflow.ExecuteActivity(ctx, CreateShipping, order).Get(ctx, &shippingResult); err != nil {
		return err
	}

	state.ShippingID = &shippingResult.ShippingID
	state.TrackingNumber = &shippingResult.TrackingNumber
	state.CompletedSteps["CreateShipping"] = true

	// Step 7: Send Order Confirmation
	logger.Info("Executing SendOrderConfirmation", "orderID", order.ID)

	confirmationInput := dto.SendOrderConfirmationRequest{
		Order:          order,
		TrackingNumber: *state.TrackingNumber,
	}
	if err := workflow.ExecuteActivity(ctx, SendOrderConfirmation, confirmationInput).Get(ctx, nil); err != nil {
		// This is not critical, log but don't fail the saga
		logger.Warn(
			"SendOrderConfirmation failed, but saga will continue",
			"error",
			err,
			"orderID",
			order.ID,
		)
	}

	state.CompletedSteps["SendOrderConfirmation"] = true

	return nil
}

// executeCompensation executes compensation activities in reverse order.
func executeCompensation(
	ctx workflow.Context,
	order *entity.Order,
	state *dto.TemporalWorkflowState,
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
	if state.CompletedSteps["CreateShipping"] && state.ShippingID != nil {
		logger.Info(
			"Compensating CreateShipping",
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

		if err := workflow.ExecuteActivity(compensationCtx, RestoreProducts, order).Get(compensationCtx, nil); err != nil {
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
	if state.CompletedSteps["ReserveProducts"] {
		logger.Info("Compensating ReserveProducts", "orderID", order.ID)

		if err := workflow.ExecuteActivity(compensationCtx, ReleaseProducts, order).Get(compensationCtx, nil); err != nil {
			logger.Error("ReleaseProducts compensation failed", "error", err, "orderID", order.ID)
		}
	}

	logger.Info("Compensation completed", "orderID", order.ID)
}
