package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// SagaState represents the persistent state of a saga execution.
type SagaState struct {
	ID               uuid.UUID              `json:"id"`
	OrderID          uuid.UUID              `json:"order_id"`
	Status           constant.SagaStatus    `json:"status"`
	CurrentStep      int                    `json:"current_step"`
	ExecutedSteps    []string               `json:"executed_steps"`
	CompensatedSteps []string               `json:"compensated_steps"`
	Data             map[string]interface{} `json:"data"`
	Error            string                 `json:"error,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
}
