// Package redisutils provides utility functions for working with Redis.
package redisutils

import (
	"fmt"

	"github.com/google/uuid"
)

// NewLockKey creates a new lock key for Redis.
func NewLockKey(idempotencyKey, userID uuid.UUID) string {
	return fmt.Sprintf("lock:%v:%v", idempotencyKey, userID)
}
