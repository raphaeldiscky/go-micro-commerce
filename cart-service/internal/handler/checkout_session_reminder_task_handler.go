// Package handler provides handlers for checkout session reminder background tasks.
package handler

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/task"
)

// CheckoutSessionReminderTaskHandler handles checkout session reminder background tasks.
type CheckoutSessionReminderTaskHandler struct {
	checkoutSessionReminderService service.CheckoutSessionReminderService
	logger                         logger.Logger
}

// NewCheckoutSessionReminderTaskHandler creates a new checkout session reminder task handler.
func NewCheckoutSessionReminderTaskHandler(
	checkoutSessionReminderService service.CheckoutSessionReminderService,
	logger logger.Logger,
) *CheckoutSessionReminderTaskHandler {
	return &CheckoutSessionReminderTaskHandler{
		checkoutSessionReminderService: checkoutSessionReminderService,
		logger:                         logger,
	}
}

// HandleCheckoutSessionReminderTask handles checkout session reminder tasks.
func (h *CheckoutSessionReminderTaskHandler) HandleCheckoutSessionReminderTask(
	ctx context.Context,
	t *asynq.Task,
) error {
	h.logger.Infof("Processing checkout session reminder task: %s", t.Type())

	// Parse the task payload
	payload, err := task.ParseCheckoutSessionReminderTask(t)
	if err != nil {
		h.logger.Errorf("Failed to parse checkout session reminder task: %v", err)
		return err
	}

	// Process the checkout session reminder
	err = h.checkoutSessionReminderService.ProcessCheckoutSessionReminder(ctx, payload)
	if err != nil {
		h.logger.Errorf(
			"Failed to process checkout session reminder for session %s: %v",
			payload.CheckoutSessionID,
			err,
		)

		return err
	}

	h.logger.Infof(
		"Successfully processed checkout session reminder for session: %s",
		payload.CheckoutSessionID,
	)

	return nil
}
