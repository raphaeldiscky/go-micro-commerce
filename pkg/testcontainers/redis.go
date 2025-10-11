package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	redisclient "github.com/redis/go-redis/v9"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	pkgredis "github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
)

const (
	defaultRedisImage    = "redis:7-alpine"
	redisStartupTimeout  = 2 * time.Minute
	redisDialTimeout     = 5 * time.Second
	redisReadTimeout     = 3 * time.Second
	redisWriteTimeout    = 3 * time.Second
	redisMinIdleConn     = 5
	redisMaxIdleConn     = 10
	redisMaxActiveConn   = 20
	redisMaxConnLifetime = 5 * time.Minute
)

// RedisConfig holds configuration for Redis testcontainer.
type RedisConfig struct {
	// Docker image (defaults to redis:7-alpine)
	Image string
	// Additional Redis arguments (e.g., ["--maxmemory", "100mb"])
	Args []string
}

// DefaultRedisConfig returns a default Redis configuration.
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Image: defaultRedisImage,
		Args:  []string{},
	}
}

// RedisContainer wraps the testcontainers Redis container.
type RedisContainer struct {
	container *redis.RedisContainer
	client    *redisclient.Client
	config    *RedisConfig
	ctx       context.Context
	addr      string
}

// NewRedisContainer creates a new Redis container instance.
func NewRedisContainer(ctx context.Context, config *RedisConfig) *RedisContainer {
	if ctx == nil {
		ctx = context.Background()
	}

	if config == nil {
		config = DefaultRedisConfig()
	}

	return &RedisContainer{
		config: config,
		ctx:    ctx,
	}
}

// Start initializes and starts the Redis container.
func (r *RedisContainer) Start() error {
	// Use default image if not specified
	image := r.config.Image
	if image == "" {
		image = defaultRedisImage
	}

	// Create container options
	opts := []testcontainers.ContainerCustomizer{
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(redisStartupTimeout),
		),
	}

	// Start the container
	container, err := redis.Run(
		r.ctx,
		image,
		opts...,
	)
	if err != nil {
		return fmt.Errorf("failed to start redis container: %w", err)
	}

	r.container = container

	// Get connection string
	connStr, err := container.ConnectionString(r.ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	// Parse address (strip redis:// prefix)
	r.addr = ParseRedisAddr(connStr)

	// Use existing pkg/redis client initialization
	redisConfig := &pkgredis.Config{
		Addr:            r.addr,
		DialTimeout:     redisDialTimeout,
		ReadTimeout:     redisReadTimeout,
		WriteTimeout:    redisWriteTimeout,
		MinIdleConn:     redisMaxIdleConn,
		MaxIdleConn:     redisMaxIdleConn,
		MaxActiveConn:   redisMaxActiveConn,
		MaxConnLifetime: redisMaxConnLifetime,
	}

	// Create logger for connection (use simple logger for tests)
	testLogger := logger.NewLogrusLogger(0)

	client, err := pkgredis.NewRedis(r.ctx, redisConfig, testLogger)
	if err != nil {
		return fmt.Errorf("failed to create redis client: %w", err)
	}

	r.client = client

	return nil
}

// GetClient returns the Redis client.
func (r *RedisContainer) GetClient() (*redisclient.Client, error) {
	if r.client == nil {
		return nil, errors.New("redis client not initialized")
	}

	return r.client, nil
}

// GetAddr returns the Redis address.
func (r *RedisContainer) GetAddr() (string, error) {
	if r.addr == "" {
		return "", errors.New("redis not started")
	}

	return r.addr, nil
}

// Terminate stops and removes the Redis container.
func (r *RedisContainer) Terminate() error {
	// Close client connection
	if r.client != nil {
		if err := r.client.Close(); err != nil {
			return fmt.Errorf("failed to close redis client: %w", err)
		}
	}

	// Terminate container
	if r.container != nil {
		if err := r.container.Terminate(r.ctx); err != nil {
			return fmt.Errorf("failed to terminate redis container: %w", err)
		}
	}

	return nil
}

// ParseRedisAddr strips the redis:// prefix from a connection string.
func ParseRedisAddr(connStr string) string {
	// Remove redis:// prefix if present
	return strings.TrimPrefix(connStr, "redis://")
}
