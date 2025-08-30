// Package redis provides utility functions for working with Redis.
package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// ClusterConfig holds the configuration for Redis Cluster.
type ClusterConfig struct {
	Addrs           []string
	Password        string
	DialTimeout     int
	ReadTimeout     int
	WriteTimeout    int
	MinIdleConn     int
	MaxIdleConn     int
	MaxActiveConn   int
	MaxConnLifetime int
}

// Config holds the configuration for standalone Redis.
type Config struct {
	Addr            string
	DialTimeout     int
	ReadTimeout     int
	WriteTimeout    int
	MinIdleConn     int
	MaxIdleConn     int
	MaxActiveConn   int
	MaxConnLifetime int
}

// NewRedisCluster initializes a Redis Cluster client and verifies the connection.
func NewRedisCluster(ctx context.Context, cfg *ClusterConfig) (*redis.ClusterClient, error) {
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           cfg.Addrs,
		Password:        cfg.Password,
		DialTimeout:     time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
		MinIdleConns:    cfg.MinIdleConn,
		MaxIdleConns:    cfg.MaxIdleConn,
		ConnMaxLifetime: time.Duration(cfg.MaxConnLifetime) * time.Minute,
		MaxActiveConns:  cfg.MaxActiveConn,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	status, err := rdb.Ping(pingCtx).Result()
	if err != nil {
		err = rdb.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close redis cluster connection: %w", err)
		}

		return nil, fmt.Errorf("failed to connect to redis cluster: %w", err)
	}

	log.Printf("redis cluster ping response: %s", status)
	log.Printf("connected to redis cluster at %v", cfg.Addrs)

	return rdb, nil
}

// NewRedis initializes a standalone Redis client and verifies the connection.
func NewRedis(ctx context.Context, cfg *Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            cfg.Addr,
		DialTimeout:     time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
		MinIdleConns:    cfg.MinIdleConn,
		MaxIdleConns:    cfg.MaxIdleConn,
		ConnMaxLifetime: time.Duration(cfg.MaxConnLifetime) * time.Minute,
		MaxActiveConns:  cfg.MaxActiveConn,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	status, err := rdb.Ping(pingCtx).Result()
	if err != nil {
		err = rdb.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close redis connection: %w", err)
		}

		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Printf("redis ping response: %s", status)
	log.Printf("connected to redis at %s", cfg.Addr)

	return rdb, nil
}
