package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// WorkflowState represents current saga execution state and status.
type WorkflowState struct {
	SagaID           uuid.UUID           `json:"saga_id"`
	OrderID          uuid.UUID           `json:"order_id"`
	Status           constant.SagaStatus `json:"status"`
	CurrentStep      int64               `json:"current_step"`
	ExecutedSteps    []string            `json:"executed_steps"`
	CompensatedSteps []string            `json:"compensated_steps"`
	Error            string              `json:"error,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	CompletedAt      *time.Time          `json:"completed_at,omitempty"`
}

// OrderSagaResponse represents an order in API responses.
type OrderSagaResponse struct {
	ID         uuid.UUID               `json:"id"`
	CustomerID uuid.UUID               `json:"customer_id"`
	Status     constant.OrderStatus    `json:"status"`
	Currency   string                  `json:"currency"`
	Items      []OrderSagaItemResponse `json:"items"`
	CreatedAt  time.Time               `json:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at"`
}

// OrderSagaItemResponse represents an order item in API responses.
type OrderSagaItemResponse struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int64     `json:"quantity"`
}
