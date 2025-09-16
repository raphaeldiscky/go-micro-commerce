// Package task provides background task definitions and payloads for order processing.
package task

import (
	"encoding/json"
	"errors"

	"github.com/hibiken/asynq"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
)

const (
	// CancelOrderTaskType is the task type for order cancellation tasks.
	CancelOrderTaskType = "order:cancel"
	// CancelOrderQueue is the queue name for order cancellation tasks.
	CancelOrderQueue = "critical"
)

// NewCancelOrderTask creates a new order cancellation task.
func NewCancelOrderTask(payload *dto.CancelOrderRequest) (*asynq.Task, error) {
	if payload == nil {
		return nil, errors.New("payload is required")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(CancelOrderTaskType, data, asynq.Queue(CancelOrderQueue)), nil
}

// ParseCancelOrderTask parses an order cancellation task payload.
func ParseCancelOrderTask(task *asynq.Task) (*dto.CancelOrderRequest, error) {
	if task.Type() != CancelOrderTaskType {
		return nil, errors.New("invalid task type")
	}

	var payload dto.CancelOrderRequest
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
