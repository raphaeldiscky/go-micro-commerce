// Package asynqutils provides utility functions for working with asynq IDs.
package asynqutils

import (
	"fmt"

	"github.com/google/uuid"
)

// GenerateTaskID creates a unique task ID for asynq tasks with correlation metadata.
func GenerateTaskID(checkoutSessionID uuid.UUID) string {
	return fmt.Sprintf(
		"checkout:reminder:%s",
		checkoutSessionID.String(),
	)
}
