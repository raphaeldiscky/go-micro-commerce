// Package redis provides utility functions for working with Redis.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// ClusterConfig holds the configuration for Redis Cluster.
type ClusterConfig struct {
	Addrs           []string
	Password        string
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MinIdleConn     int
	MaxIdleConn     int
	MaxActiveConn   int
	MaxConnLifetime time.Duration
}

// Config holds the configuration for standalone Redis.
type Config struct {
	Addr            string
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MinIdleConn     int
	MaxIdleConn     int
	MaxActiveConn   int
	MaxConnLifetime time.Duration
}

// NewRedisCluster initializes a Redis Cluster client and verifies the connection.
func NewRedisCluster(
	ctx context.Context,
	cfg *ClusterConfig,
	appLogger logger.Logger,
) (*redis.ClusterClient, error) {
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           cfg.Addrs,
		Password:        cfg.Password,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		MinIdleConns:    cfg.MinIdleConn,
		MaxIdleConns:    cfg.MaxIdleConn,
		ConnMaxLifetime: cfg.MaxConnLifetime,
		MaxActiveConns:  cfg.MaxActiveConn,
	})

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
	defer cancel()

	status, err := rdb.Ping(pingCtx).Result()
	if err != nil {
		originalErr := err

		closeErr := rdb.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("failed to close redis cluster connection: %w", closeErr)
		}

		return nil, fmt.Errorf("failed to connect to redis cluster: %w", originalErr)
	}

	appLogger.Printf("redis cluster ping response: %s", status)
	appLogger.Printf("connected to redis cluster at %v", cfg.Addrs)

	return rdb, nil
}

// NewRedis initializes a standalone Redis client and verifies the connection.
func NewRedis(ctx context.Context, cfg *Config, appLogger logger.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            cfg.Addr,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		MinIdleConns:    cfg.MinIdleConn,
		MaxIdleConns:    cfg.MaxIdleConn,
		ConnMaxLifetime: cfg.MaxConnLifetime,
		MaxActiveConns:  cfg.MaxActiveConn,
	})

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
	defer cancel()

	status, err := rdb.Ping(pingCtx).Result()
	if err != nil {
		originalErr := err

		closeErr := rdb.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("failed to close redis connection: %w", closeErr)
		}

		return nil, fmt.Errorf("failed to connect to redis: %w", originalErr)
	}

	appLogger.Printf("redis ping response: %s", status)
	appLogger.Printf("connected to redis at %s", cfg.Addr)

	return rdb, nil
}
