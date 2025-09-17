package temporal

import (
	"github.com/google/uuid"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
)

// PaymentReminderWorkflow sends payment reminders to customers using Temporal timers (matching saga timing).
func PaymentReminderWorkflow(
	ctx workflow.Context,
	req dto.PaymentReminderWorkflowRequest,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Payment Reminder Workflow", "orderID", req.OrderID)

	// Configure activity options for payment reminders
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: constant.PaymentReminderTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    constant.PaymentReminderInitialInterval,
			BackoffCoefficient: constant.PaymentReminderBackoffCoefficient,
			MaximumInterval:    constant.PaymentReminderMaxInterval,
			MaximumAttempts:    constant.PaymentReminderMaxAttempts,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Create payment cancellation signal channel
	paymentCancelledSignal := workflow.GetSignalChannel(ctx, "payment-cancelled")

	// Setup reminder requests
	firstReminderRequest := createReminderRequest(req, constant.FirstReminderSequence)
	secondReminderRequest := createReminderRequest(req, constant.SecondReminderSequence)
	expirationRequest := createExpirationRequest(req, firstReminderRequest.CorrelationID)

	// Setup timers
	firstReminderTimer := workflow.NewTimer(ctx, constant.FirstPaymentReminderMinutes)
	secondReminderTimer := workflow.NewTimer(ctx, constant.SecondPaymentReminderMinutes)
	orderExpirationTimer := workflow.NewTimer(ctx, constant.CancelOrderDelayMinutes)

	// Process first reminder
	if !processReminderStage(ctx, logger, req.OrderID, firstReminderTimer, paymentCancelledSignal,
		constant.SendPaymentReminderActivity, firstReminderRequest) {
		return nil // Payment was cancelled
	}

	// Process second reminder
	if !processReminderStage(ctx, logger, req.OrderID, secondReminderTimer, paymentCancelledSignal,
		constant.SendPaymentReminderActivity, secondReminderRequest) {
		return nil // Payment was cancelled
	}

	// Process order expiration
	processExpirationStage(
		ctx,
		logger,
		req.OrderID,
		orderExpirationTimer,
		paymentCancelledSignal,
		expirationRequest,
	)

	logger.Info("Payment Reminder Workflow completed", "orderID", req.OrderID)

	return nil
}

func createReminderRequest(
	req dto.PaymentReminderWorkflowRequest,
	reminderCount int,
) dto.PaymentReminderRequest {
	return dto.PaymentReminderRequest{
		OrderID:          req.OrderID,
		CorrelationID:    uuid.New(),
		CustomerEmail:    req.CustomerEmail,
		PaymentID:        req.PaymentID,
		TotalPrice:       req.TotalPrice,
		Currency:         req.Currency,
		ReminderCount:    reminderCount,
		ReservedProducts: nil, // Will be set by the activity
	}
}

func createExpirationRequest(
	req dto.PaymentReminderWorkflowRequest,
	correlationID uuid.UUID,
) dto.ExpireOrderPaymentRequest {
	return dto.ExpireOrderPaymentRequest{
		CustomerID:     uuid.Nil, // Will be set by the activity
		CustomerEmail:  req.CustomerEmail,
		OrderID:        req.OrderID,
		CorrelationID:  correlationID,
		IdempotencyKey: uuid.New(),
	}
}

func processReminderStage(
	ctx workflow.Context,
	logger log.Logger,
	orderID uuid.UUID,
	timer workflow.Future,
	cancelSignal workflow.ReceiveChannel,
	activityName constant.WorkflowStep,
	request dto.PaymentReminderRequest,
) bool {
	selector := workflow.NewSelector(ctx)

	selector.AddFuture(timer, func(_ workflow.Future) {
		handleReminderTimer(ctx, logger, orderID, activityName, request)
	})

	selector.AddReceive(cancelSignal, func(ch workflow.ReceiveChannel, _ bool) {
		ch.ReceiveAsync(nil) // Clear the signal
		logger.Info("Payment cancelled signal received, stopping workflow", "orderID", orderID)
	})

	selector.Select(ctx)

	// Check if we should continue (false if cancelled)
	var signalValue interface{}

	ok := cancelSignal.ReceiveAsync(&signalValue)

	return !ok
}

func processExpirationStage(
	ctx workflow.Context,
	logger log.Logger,
	orderID uuid.UUID,
	timer workflow.Future,
	cancelSignal workflow.ReceiveChannel,
	request dto.ExpireOrderPaymentRequest,
) {
	selector := workflow.NewSelector(ctx)

	selector.AddFuture(timer, func(_ workflow.Future) {
		handleExpirationTimer(ctx, logger, orderID, request)
	})

	selector.AddReceive(cancelSignal, func(ch workflow.ReceiveChannel, _ bool) {
		ch.ReceiveAsync(nil) // Clear the signal
		logger.Info("Payment cancelled signal received, stopping workflow", "orderID", orderID)
	})

	selector.Select(ctx)
}

func handleReminderTimer(
	ctx workflow.Context,
	logger log.Logger,
	orderID uuid.UUID,
	activityName constant.WorkflowStep,
	request dto.PaymentReminderRequest,
) {
	logger.Info(
		"Payment reminder timer fired",
		"orderID",
		orderID,
		"reminderCount",
		request.ReminderCount,
	)

	paymentReceived := checkPaymentStatus(ctx, logger, orderID)
	if !paymentReceived {
		sendReminder(ctx, logger, orderID, activityName, request)
	} else {
		logger.Info("Payment received, skipping reminder", "orderID", orderID, "reminderCount", request.ReminderCount)
	}
}

func handleExpirationTimer(
	ctx workflow.Context,
	logger log.Logger,
	orderID uuid.UUID,
	request dto.ExpireOrderPaymentRequest,
) {
	logger.Info("Order expiration timer fired", "orderID", orderID)

	paymentReceived := checkPaymentStatus(ctx, logger, orderID)
	if !paymentReceived {
		expireOrder(ctx, logger, orderID, request)
	} else {
		logger.Info("Payment received, skipping order expiration", "orderID", orderID)
	}
}

func checkPaymentStatus(ctx workflow.Context, logger log.Logger, orderID uuid.UUID) bool {
	var paymentReceived bool

	err := workflow.ExecuteActivity(ctx, string(constant.CheckPaymentStatusActivity), orderID).
		Get(ctx, &paymentReceived)
	if err != nil {
		logger.Warn("Failed to check payment status", "orderID", orderID, "error", err)
	}

	return paymentReceived
}

func sendReminder(
	ctx workflow.Context,
	logger log.Logger,
	orderID uuid.UUID,
	activityName constant.WorkflowStep,
	request dto.PaymentReminderRequest,
) {
	err := workflow.ExecuteActivity(ctx, string(activityName), request).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to send payment reminder", "orderID", orderID, "error", err)
	} else {
		logger.Info("Successfully sent payment reminder", "orderID", orderID, "reminderCount", request.ReminderCount)
	}
}

func expireOrder(
	ctx workflow.Context,
	logger log.Logger,
	orderID uuid.UUID,
	request dto.ExpireOrderPaymentRequest,
) {
	err := workflow.ExecuteActivity(ctx, string(constant.ExpireOrderPaymentActivity), request).
		Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to expire order payment", "orderID", orderID, "error", err)
	} else {
		logger.Info("Successfully expired order payment", "orderID", orderID)
	}
}
