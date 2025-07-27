package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements a Redis-based cache.
type RedisCache struct {
	client     *redis.Client
	keyPrefix  string
	defaultTTL time.Duration
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Host       string
	Port       int
	Password   string
	DB         int
	KeyPrefix  string
	DefaultTTL time.Duration
}

// NewRedisCache creates a new Redis cache instance.
func NewRedisCache(config RedisConfig) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	return &RedisCache{
		client:     rdb,
		keyPrefix:  config.KeyPrefix,
		defaultTTL: config.DefaultTTL,
	}
}

// Set stores a value in the cache with the default TTL.
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	return c.SetWithTTL(ctx, key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL.
func (c *RedisCache) SetWithTTL(
	ctx context.Context,
	key string,
	value interface{},
	ttl time.Duration,
) error {
	fullKey := c.buildKey(key)

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = c.client.Set(ctx, fullKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	return nil
}

// Get retrieves a value from the cache.
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := c.buildKey(key)

	data, err := c.client.Get(ctx, fullKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}

		return fmt.Errorf("failed to get cache value: %w", err)
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return nil
}

// Delete removes a value from the cache.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.buildKey(key)

	err := c.client.Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache value: %w", err)
	}

	return nil
}

// Exists checks if a key exists in the cache.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.buildKey(key)

	count, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache key existence: %w", err)
	}

	return count > 0, nil
}

// Clear removes all keys with the configured prefix.
func (c *RedisCache) Clear(ctx context.Context) error {
	pattern := c.buildKey("*")

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	err = c.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	return nil
}

// Ping checks if Redis is reachable.
func (c *RedisCache) Ping(ctx context.Context) error {
	err := c.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	return nil
}

// Close closes the Redis connection.
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// buildKey builds the full cache key with prefix.
func (c *RedisCache) buildKey(key string) string {
	if c.keyPrefix != "" {
		return fmt.Sprintf("%s:%s", c.keyPrefix, key)
	}

	return key
}

// ErrCacheMiss is returned when a cache key is not found.
var ErrCacheMiss = fmt.Errorf("cache miss")
