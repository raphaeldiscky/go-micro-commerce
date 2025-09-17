package temporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
)

// PaymentReminderWorkflow sends payment reminders to customers with escalating urgency.
func PaymentReminderWorkflow(
	ctx workflow.Context,
	req dto.PaymentReminderWorkflowRequest,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Payment Reminder Workflow", "orderID", req.OrderID)

	// Configure activity options with shorter timeout for reminders
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

	// Keep track of reminder count to support escalation
	reminderCount := 0

	// Use hardcoded config for payment reminders: 15min, 40min, 55min
	maxReminders := 3
	reminderTimes := constant.GetPaymentReminderExecutionTimes()

	for reminderCount < maxReminders {
		reminderCount++

		logger.Info(
			"Sending payment reminder",
			"orderID", req.OrderID,
			"reminderNumber", reminderCount,
			"maxReminders", maxReminders,
		)

		// Prepare reminder request
		reminderRequest := dto.PaymentReminderRequest{
			OrderID:       req.OrderID,
			CustomerEmail: req.CustomerEmail,
			PaymentID:     req.PaymentID,
			TotalPrice:    req.TotalPrice,
			Currency:      req.Currency,
			ReminderCount: reminderCount,
		}

		// Send reminder notification
		err := workflow.ExecuteActivity(ctx, string(constant.SendPaymentReminderActivity), reminderRequest).
			Get(ctx, nil)
		if err != nil {
			logger.Error(
				"Failed to send payment reminder",
				"orderID", req.OrderID,
				"reminderNumber", reminderCount,
				"error", err,
			)
			// Continue to next reminder even if this one fails
		} else {
			logger.Info(
				"Successfully sent payment reminder",
				"orderID", req.OrderID,
				"reminderNumber", reminderCount,
			)
		}

		// If this was the last reminder, exit
		if reminderCount >= maxReminders {
			break
		}

		// Calculate next reminder time
		var nextInterval time.Duration
		if reminderCount <= len(reminderTimes) {
			nextInterval = reminderTimes[reminderCount-1]
		} else {
			// Use last interval if we've exceeded the configured intervals
			nextInterval = reminderTimes[len(reminderTimes)-1]
		}

		logger.Info(
			"Waiting for next reminder interval",
			"orderID", req.OrderID,
			"nextInterval", nextInterval,
		)

		// Wait for the next reminder interval
		err = workflow.Sleep(ctx, nextInterval)
		if err != nil {
			logger.Error(
				"Workflow sleep interrupted",
				"orderID", req.OrderID,
				"error", err,
			)

			break
		}

		// Check if payment has been received (activity will handle this check)
		var paymentReceived bool

		err = workflow.ExecuteActivity(ctx, string(constant.CheckPaymentStatusActivity), req.OrderID).
			Get(ctx, &paymentReceived)
		if err != nil {
			logger.Warn(
				"Failed to check payment status",
				"orderID", req.OrderID,
				"error", err,
			)
			// Continue with reminder in case of check failure
		} else if paymentReceived {
			logger.Info(
				"Payment received, stopping reminder workflow",
				"orderID", req.OrderID,
			)

			break
		}
	}

	if reminderCount >= maxReminders {
		logger.Warn(
			"Maximum reminder attempts reached",
			"orderID", req.OrderID,
			"reminderCount", reminderCount,
		)

		// Execute final escalation reminder (reminder count will be at max)
		finalReminderRequest := dto.PaymentReminderRequest{
			OrderID:       req.OrderID,
			CustomerEmail: req.CustomerEmail,
			PaymentID:     req.PaymentID,
			TotalPrice:    req.TotalPrice,
			Currency:      req.Currency,
			ReminderCount: reminderCount,
		}

		err := workflow.ExecuteActivity(ctx, string(constant.SendPaymentReminderActivity), finalReminderRequest).
			Get(ctx, nil)
		if err != nil {
			logger.Error(
				"Failed to send final payment reminder",
				"orderID", req.OrderID,
				"error", err,
			)
		}
	}

	logger.Info("Payment Reminder Workflow completed", "orderID", req.OrderID)

	return nil
}
