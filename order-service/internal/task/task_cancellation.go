package task

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// CancellationHelper provides utilities for cancelling order-related tasks.
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

// CancelPaymentReminderTasksByIDs cancels payment reminder tasks by their task IDs.
func (h *CancellationHelper) CancelPaymentReminderTasksByIDs(
	ctx context.Context,
	taskIDs []string,
) error {
	if len(taskIDs) == 0 {
		h.logger.Info("No task IDs provided for payment reminder cancellation")
		return nil
	}

	h.logger.Infof("Cancelling %d payment reminder tasks: %v", len(taskIDs), taskIDs)

	for _, taskID := range taskIDs {
		if err := h.cancellationService.CancelTask(ctx, PaymentReminderQueue, taskID); err != nil {
			h.logger.Errorf("Failed to cancel payment reminder task %s: %v", taskID, err)
			// Continue with other tasks even if one fails
		} else {
			h.logger.Infof("Successfully cancelled payment reminder task: %s", taskID)
		}
	}

	return nil
}
