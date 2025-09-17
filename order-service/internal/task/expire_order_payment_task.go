// Package task provides background task definitions and payloads for order processing.
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
	// ExpireOrderPaymentTaskType is the task type for order cancellation tasks.
	ExpireOrderPaymentTaskType = "order:payment_expired"
	// ExpireOrderPaymentQueue is the queue name for order cancellation tasks.
	ExpireOrderPaymentQueue = "critical"
)

// NewExpireOrderPaymentTask creates a new order cancellation task with correlation metadata.
func NewExpireOrderPaymentTask(payload dto.ExpireOrderPaymentRequest) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Add task options with correlation metadata for cleanup
	opts := []asynq.Option{
		asynq.Queue(ExpireOrderPaymentQueue),
		asynq.TaskID(sagautils.GenerateCancelTaskID(payload.OrderID, payload.CorrelationID)),
		asynq.Retention(constant.TaskRetentionHours), // Keep completed tasks for debugging
	}

	return asynq.NewTask(ExpireOrderPaymentTaskType, data, opts...), nil
}

// ParseExpireOrderPaymentTask parses an order cancellation task payload.
func ParseExpireOrderPaymentTask(task *asynq.Task) (*dto.ExpireOrderPaymentRequest, error) {
	if task.Type() != ExpireOrderPaymentTaskType {
		return nil, errors.New("invalid task type")
	}

	var payload dto.ExpireOrderPaymentRequest
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
