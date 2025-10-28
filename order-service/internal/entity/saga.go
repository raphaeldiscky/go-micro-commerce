package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
)

// SagaState represents the persistent state of a saga execution.
type SagaState struct {
	ID               uuid.UUID             `json:"id"`
	WorkflowName     constant.WorkflowName `json:"workflow_name"`
	OrderID          uuid.UUID             `json:"order_id"`
	Status           constant.SagaStatus   `json:"status"`
	CurrentStep      int64                 `json:"current_step"`
	ExecutedSteps    []string              `json:"executed_steps"`
	CompensatedSteps []string              `json:"compensated_steps"`
	Data             map[string]any        `json:"data"`
	Error            string                `json:"error,omitempty"`
	Version          int64                 `json:"version"`
	RetryCount       int64                 `json:"retry_count"`
	LastRetryAt      *time.Time            `json:"last_retry_at,omitempty"`
	TimeoutAt        *time.Time            `json:"timeout_at,omitempty"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	CompletedAt      *time.Time            `json:"completed_at,omitempty"`
}

// CanRetry checks if saga can be retried based on retry count and timeout.
func (s *SagaState) CanRetry(maxRetries int64, maxAge time.Duration) bool {
	if s.RetryCount >= maxRetries {
		return false
	}

	if s.TimeoutAt != nil && time.Now().After(*s.TimeoutAt) {
		return false
	}

	if time.Since(s.CreatedAt) > maxAge {
		return false
	}

	return true
}

// IncrementRetry increments the retry count and updates last retry time.
func (s *SagaState) IncrementRetry() {
	s.RetryCount++
	now := time.Now().UTC()
	s.LastRetryAt = &now
	s.UpdatedAt = now
}

// SetTimeout sets the timeout for the saga.
func (s *SagaState) SetTimeout(timeout time.Duration) {
	timeoutAt := s.CreatedAt.Add(timeout)
	s.TimeoutAt = &timeoutAt
}
