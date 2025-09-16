package handler

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/task"
)

// PaymentReminderTaskHandler handles payment reminder background tasks.
type PaymentReminderTaskHandler struct {
	paymentReminderService service.PaymentReminderTaskService
	logger                 logger.Logger
}

// NewPaymentReminderTaskHandler creates a new payment reminder task handler.
func NewPaymentReminderTaskHandler(
	paymentReminderService service.PaymentReminderTaskService,
	logger logger.Logger,
) *PaymentReminderTaskHandler {
	return &PaymentReminderTaskHandler{
		paymentReminderService: paymentReminderService,
		logger:                 logger,
	}
}

// HandlePaymentReminderTask handles payment reminder tasks.
func (h *PaymentReminderTaskHandler) HandlePaymentReminderTask(
	ctx context.Context,
	t *asynq.Task,
) error {
	h.logger.Infof("Processing payment reminder task: %s", t.Type())

	// Parse the task payload
	payload, err := task.ParsePaymentReminderTask(t)
	if err != nil {
		h.logger.Errorf("Failed to parse payment reminder task: %v", err)
		return err
	}

	// Process the payment reminder
	err = h.paymentReminderService.ProcessPaymentReminder(ctx, payload)
	if err != nil {
		h.logger.Errorf("Failed to process payment reminder for order %s: %v", payload.OrderID, err)
		return err
	}

	h.logger.Infof("Successfully processed payment reminder for order: %s", payload.OrderID)

	return nil
}

// HandleCancelOrderTask handles order cancellation tasks.
func (h *PaymentReminderTaskHandler) HandleCancelOrderTask(
	ctx context.Context,
	t *asynq.Task,
) error {
	h.logger.Infof("Processing cancel order task: %s", t.Type())

	// Parse the task payload
	payload, err := task.ParseCancelOrderTask(t)
	if err != nil {
		h.logger.Errorf("Failed to parse cancel order task: %v", err)
		return err
	}

	// Process the order cancellation
	err = h.paymentReminderService.ProcessOrderCancellation(ctx, payload)
	if err != nil {
		h.logger.Errorf(
			"Failed to process order cancellation for order %s: %v",
			payload.OrderID,
			err,
		)

		return err
	}

	h.logger.Infof("Successfully processed order cancellation for order: %s", payload.OrderID)

	return nil
}
