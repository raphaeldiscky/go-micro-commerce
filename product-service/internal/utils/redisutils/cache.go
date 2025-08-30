// Package redisutils provides utility functions for working with Redis.
package redisutils

import (
	"fmt"

	"github.com/google/uuid"
)

// NewCacheListProductsKey creates a new cache key for listing products.
func NewCacheListProductsKey(page, limit int64) string {
	return fmt.Sprintf("cache:product:list:page-%v:limit-%v", page, limit)
}

// NewCacheProductKey creates a new cache key for individual product.
func NewCacheProductKey(productID uuid.UUID) string {
	return fmt.Sprintf("cache:product:%v", productID)
}

// NewCacheListProductsPatternKey creates a new cache pattern key for listing products.
func NewCacheListProductsPatternKey() string {
	return "cache:product:list:*"
}
