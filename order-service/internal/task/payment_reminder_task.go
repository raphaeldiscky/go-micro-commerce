package task

import (
	"encoding/json"
	"errors"

	"github.com/hibiken/asynq"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/utils/sagautils"
)

const (
	// PaymentReminderTaskType is the task type for payment reminder tasks.
	PaymentReminderTaskType = "payment:reminder"
	// PaymentReminderQueue is the queue name for payment reminder tasks.
	PaymentReminderQueue = "critical"
)

// NewPaymentReminderTask creates a new payment reminder task with correlation metadata.
func NewPaymentReminderTask(payload *dto.PaymentReminderRequest) (*asynq.Task, error) {
	if payload == nil {
		return nil, errors.New("payload is required")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Add task options with correlation metadata for cleanup
	opts := []asynq.Option{
		asynq.Queue(PaymentReminderQueue),
		asynq.TaskID(
			sagautils.GenerateTaskID(payload.OrderID, payload.CorrelationID, payload.ReminderCount),
		),
		asynq.Retention(constant.TaskRetentionHours), // Keep completed tasks for debugging
	}

	return asynq.NewTask(PaymentReminderTaskType, data, opts...), nil
}

// ParsePaymentReminderTask parses a payment reminder task payload.
func ParsePaymentReminderTask(task *asynq.Task) (*dto.PaymentReminderRequest, error) {
	if task.Type() != PaymentReminderTaskType {
		return nil, errors.New("invalid task type")
	}

	var payload dto.PaymentReminderRequest
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
