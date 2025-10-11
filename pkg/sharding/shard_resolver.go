// Package sharding provides sharding utilities for the application.
package sharding

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	ch "github.com/ArchishmanSengupta/consistent-hashing"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// ShardResolver provides consistent hashing for shard distribution.
type ShardResolver struct {
	consistent *ch.ConsistentHashing
	logger     logger.Logger
}

// NewShardResolver creates a new shard resolver with consistent hashing.
func NewShardResolver(config Config, appLogger logger.Logger) (*ShardResolver, error) {
	chConfig := ch.Config{
		ReplicationFactor: config.ReplicationFactor,
		LoadFactor:        config.LoadFactor,
	}

	consistent, err := ch.NewWithConfig(chConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consistent hashing: %w", err)
	}

	resolver := &ShardResolver{
		consistent: consistent,
		logger:     appLogger,
	}

	// Initialize all shards
	ctx := context.Background()

	for i := range config.ShardCount {
		shardName := shardKey(i)
		if err = resolver.consistent.Add(ctx, shardName); err != nil {
			return nil, fmt.Errorf("failed to add shard %s: %w", shardName, err)
		}
	}

	return resolver, nil
}

// GetShardForUser returns the shard ID for a given user UUID.
func (s *ShardResolver) GetShardForUser(userID uuid.UUID) (int, error) {
	ctx := context.Background()

	shardName, err := s.consistent.GetLeast(ctx, userID.String())
	if err != nil {
		return 0, fmt.Errorf("failed to get shard for user %s: %w", userID, err)
	}

	// Increase load for the assigned shard
	if err = s.consistent.IncreaseLoad(ctx, shardName); err != nil {
		// Log this error but don't fail the operation
		s.logger.Errorf("Warning: failed to increase load for shard %s: %v\n", shardName, err)
	}

	return parseShardKey(shardName), nil
}

// GetShardForString returns the shard ID for any string identifier.
func (s *ShardResolver) GetShardForString(id string) (int, error) {
	ctx := context.Background()

	shardName, err := s.consistent.GetLeast(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("failed to get shard for ID %s: %w", id, err)
	}

	// Increase load for the assigned shard
	if err = s.consistent.IncreaseLoad(ctx, shardName); err != nil {
		s.logger.Errorf("Warning: failed to increase load for shard %s: %v\n", shardName, err)
	}

	return parseShardKey(shardName), nil
}

// GetShardLoads returns current load distribution across shards.
func (s *ShardResolver) GetShardLoads() map[string]int64 {
	return s.consistent.GetLoads()
}

// GetAllShards returns list of all active shards.
func (s *ShardResolver) GetAllShards() []string {
	return s.consistent.Hosts()
}

// RemoveShard removes a shard from the ring (for maintenance).
func (s *ShardResolver) RemoveShard(shardID int) error {
	ctx := context.Background()
	return s.consistent.Remove(ctx, shardKey(shardID))
}

// AddShard adds a shard back to the ring.
func (s *ShardResolver) AddShard(shardID int) error {
	ctx := context.Background()
	return s.consistent.Add(ctx, shardKey(shardID))
}

func shardKey(shardID int) string {
	return fmt.Sprintf("shard-%d", shardID)
}

func parseShardKey(shard string) int {
	var shardID int

	_, err := fmt.Sscanf(shard, "shard-%d", &shardID)
	if err != nil {
		return -1
	}

	return shardID
}
