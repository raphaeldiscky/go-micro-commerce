// Package task provides background task definitions and payloads for cart processing.
package task

import (
	"errors"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/utils/asynqutils"
)

const (
	// CheckoutSessionReminderTaskType is the task type for checkout session reminder tasks.
	CheckoutSessionReminderTaskType = "checkout:session:reminder"
	// CheckoutSessionReminderQueue is the queue name for checkout session reminder tasks.
	CheckoutSessionReminderQueue = "default"
)

// NewCheckoutSessionReminderTask creates a new checkout session reminder task with metadata.
func NewCheckoutSessionReminderTask(
	payload *dto.CheckoutSessionReminderRequest,
) (*asynq.Task, error) {
	if payload == nil {
		return nil, errors.New("payload is required")
	}

	data, err := sonic.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Add task options with correlation metadata for cleanup
	opts := []asynq.Option{
		asynq.Queue(CheckoutSessionReminderQueue),
		asynq.TaskID(
			asynqutils.GenerateTaskID(
				payload.CheckoutSessionID,
			),
		),
		asynq.Retention(constant.TaskRetentionHours), // Keep completed tasks for debugging
	}

	return asynq.NewTask(CheckoutSessionReminderTaskType, data, opts...), nil
}

// ParseCheckoutSessionReminderTask parses a checkout session reminder task payload.
func ParseCheckoutSessionReminderTask(
	task *asynq.Task,
) (*dto.CheckoutSessionReminderRequest, error) {
	if task.Type() != CheckoutSessionReminderTaskType {
		return nil, errors.New("invalid task type")
	}

	var payload dto.CheckoutSessionReminderRequest
	if err := sonic.Unmarshal(task.Payload(), &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
