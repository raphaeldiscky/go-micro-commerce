package testcontainers_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/testcontainers"
)

func TestRedisContainer_StartAndConnect(t *testing.T) {
	ctx := context.Background()

	// Create Redis container with default config
	config := testcontainers.DefaultRedisConfig()
	redisContainer := testcontainers.NewRedisContainer(ctx, config)

	// Start container
	err := redisContainer.Start()
	require.NoError(t, err, "Failed to start Redis container")

	defer func() {
		err = redisContainer.Terminate()
		require.NoError(t, err, "Failed to terminate Redis container")
	}()

	// Get client
	client, err := redisContainer.GetClient()
	require.NoError(t, err, "Failed to get Redis client")
	require.NotNil(t, client, "Redis client should not be nil")

	// Test connection with PING
	pong, err := client.Ping(ctx).Result()
	require.NoError(t, err, "Failed to ping Redis")
	require.Equal(t, "PONG", pong, "Ping response should be PONG")
}

func TestRedisContainer_GetAddr(t *testing.T) {
	ctx := context.Background()

	// Create and start container
	config := testcontainers.DefaultRedisConfig()
	redisContainer := testcontainers.NewRedisContainer(ctx, config)

	err := redisContainer.Start()
	require.NoError(t, err, "Failed to start Redis container")

	defer func() {
		err = redisContainer.Terminate()
		require.NoError(t, err, "Failed to terminate Redis container")
	}()

	// Get address
	addr, err := redisContainer.GetAddr()
	require.NoError(t, err, "Failed to get Redis address")
	require.NotEmpty(t, addr, "Redis address should not be empty")
	require.NotContains(t, addr, "redis://", "Address should not contain redis:// prefix")
}

func TestRedisContainer_SetAndGet(t *testing.T) {
	ctx := context.Background()

	// Create and start container
	config := testcontainers.DefaultRedisConfig()
	redisContainer := testcontainers.NewRedisContainer(ctx, config)

	err := redisContainer.Start()
	require.NoError(t, err, "Failed to start Redis container")

	defer func() {
		err = redisContainer.Terminate()
		require.NoError(t, err, "Failed to terminate Redis container")
	}()

	// Get client
	client, err := redisContainer.GetClient()
	require.NoError(t, err, "Failed to get Redis client")

	// Set a key
	err = client.Set(ctx, "test_key", "test_value", 0).Err()
	require.NoError(t, err, "Failed to set Redis key")

	// Get the key
	value, err := client.Get(ctx, "test_key").Result()
	require.NoError(t, err, "Failed to get Redis key")
	require.Equal(t, "test_value", value, "Value should match")
}

func TestRedisContainer_ExpireKey(t *testing.T) {
	ctx := context.Background()

	// Create and start container
	config := testcontainers.DefaultRedisConfig()
	redisContainer := testcontainers.NewRedisContainer(ctx, config)

	err := redisContainer.Start()
	require.NoError(t, err, "Failed to start Redis container")

	defer func() {
		err = redisContainer.Terminate()
		require.NoError(t, err, "Failed to terminate Redis container")
	}()

	// Get client
	client, err := redisContainer.GetClient()
	require.NoError(t, err, "Failed to get Redis client")

	// Set a key with expiration
	err = client.Set(ctx, "expire_key", "expire_value", 1*time.Second).Err()
	require.NoError(t, err, "Failed to set Redis key with expiration")

	// Key should exist immediately
	value, err := client.Get(ctx, "expire_key").Result()
	require.NoError(t, err, "Failed to get Redis key")
	require.Equal(t, "expire_value", value, "Value should match")

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Key should not exist after expiration
	_, err = client.Get(ctx, "expire_key").Result()
	require.Error(t, err, "Key should not exist after expiration")
	require.Contains(t, err.Error(), "redis: nil", "Error should indicate key not found")
}

func TestRedisContainer_HashOperations(t *testing.T) {
	ctx := context.Background()

	// Create and start container
	config := testcontainers.DefaultRedisConfig()
	redisContainer := testcontainers.NewRedisContainer(ctx, config)

	err := redisContainer.Start()
	require.NoError(t, err, "Failed to start Redis container")

	defer func() {
		err = redisContainer.Terminate()
		require.NoError(t, err, "Failed to terminate Redis container")
	}()

	// Get client
	client, err := redisContainer.GetClient()
	require.NoError(t, err, "Failed to get Redis client")

	// Set hash fields
	err = client.HSet(ctx, "test_hash", "field1", "value1", "field2", "value2").Err()
	require.NoError(t, err, "Failed to set hash fields")

	// Get hash field
	value, err := client.HGet(ctx, "test_hash", "field1").Result()
	require.NoError(t, err, "Failed to get hash field")
	require.Equal(t, "value1", value, "Value should match")

	// Get all hash fields
	allFields, err := client.HGetAll(ctx, "test_hash").Result()
	require.NoError(t, err, "Failed to get all hash fields")
	require.Len(t, allFields, 2, "Should have 2 fields")
	require.Equal(t, "value1", allFields["field1"], "Field1 should match")
	require.Equal(t, "value2", allFields["field2"], "Field2 should match")
}

func TestRedisContainer_ListOperations(t *testing.T) {
	ctx := context.Background()

	// Create and start container
	config := testcontainers.DefaultRedisConfig()
	redisContainer := testcontainers.NewRedisContainer(ctx, config)

	err := redisContainer.Start()
	require.NoError(t, err, "Failed to start Redis container")

	defer func() {
		err = redisContainer.Terminate()
		require.NoError(t, err, "Failed to terminate Redis container")
	}()

	// Get client
	client, err := redisContainer.GetClient()
	require.NoError(t, err, "Failed to get Redis client")

	// Push items to list
	err = client.RPush(ctx, "test_list", "item1", "item2", "item3").Err()
	require.NoError(t, err, "Failed to push items to list")

	// Get list length
	length, err := client.LLen(ctx, "test_list").Result()
	require.NoError(t, err, "Failed to get list length")
	require.Equal(t, int64(3), length, "List length should be 3")

	// Get list range
	items, err := client.LRange(ctx, "test_list", 0, -1).Result()
	require.NoError(t, err, "Failed to get list range")
	require.Len(t, items, 3, "Should have 3 items")
	require.Equal(t, []string{"item1", "item2", "item3"}, items, "Items should match")
}

func TestRedisContainer_ErrorBeforeStart(t *testing.T) {
	ctx := context.Background()

	// Create container without starting
	config := testcontainers.DefaultRedisConfig()
	redisContainer := testcontainers.NewRedisContainer(ctx, config)

	// Should return error when getting client before start
	client, err := redisContainer.GetClient()
	require.Error(t, err, "Should return error when getting client before start")
	require.Nil(t, client, "Client should be nil before start")

	// Should return error when getting address before start
	addr, err := redisContainer.GetAddr()
	require.Error(t, err, "Should return error when getting address before start")
	require.Empty(t, addr, "Address should be empty before start")
}

func TestParseRedisAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with redis:// prefix",
			input:    "redis://localhost:6379",
			expected: "localhost:6379",
		},
		{
			name:     "without prefix",
			input:    "localhost:6379",
			expected: "localhost:6379",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testcontainers.ParseRedisAddr(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
