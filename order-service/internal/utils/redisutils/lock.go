// Package redisutils provides utility functions for working with Redis.
package redisutils

import (
	"fmt"
)

// NewLockKey creates a new lock key for Redis.
func NewLockKey(requestID string, userID int64) string {
	return fmt.Sprintf("lock:%v-%v", requestID, userID)
}
