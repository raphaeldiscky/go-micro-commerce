package task

import (
	"context"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/utils/asynqutils"
)

// CancellationHelper provides utilities for cancelling checkout session reminder tasks.
type CancellationHelper struct {
	cancellationService asynq.TaskCancellationService
	logger              logger.Logger
}

// NewCancellationHelper creates a new task cancellation helper.
func NewCancellationHelper(
	cancellationService asynq.TaskCancellationService,
	logger logger.Logger,
) *CancellationHelper {
	return &CancellationHelper{
		cancellationService: cancellationService,
		logger:              logger,
	}
}

// CancelCheckoutSessionReminderTask cancels a checkout session reminder task by checkout session ID.
func (h *CancellationHelper) CancelCheckoutSessionReminderTask(
	ctx context.Context,
	checkoutSessionID uuid.UUID,
) error {
	taskID := asynqutils.GenerateTaskID(checkoutSessionID)

	h.logger.Infof("Cancelling checkout session reminder task: %s", taskID)

	if err := h.cancellationService.CancelTask(ctx, CheckoutSessionReminderQueue, taskID); err != nil {
		h.logger.Errorf("Failed to cancel checkout session reminder task %s: %v", taskID, err)
		return err
	}

	h.logger.Infof("Successfully cancelled checkout session reminder task: %s", taskID)

	return nil
}
