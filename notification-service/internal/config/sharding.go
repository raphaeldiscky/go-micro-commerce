package config

import (
	"github.com/spf13/viper"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// ShardingConfig holds sharding configuration for consistent hashing.
type ShardingConfig struct {
	ShardCount        int     `mapstructure:"SHARDING_SHARD_COUNT"`
	ReplicationFactor int     `mapstructure:"SHARDING_REPLICATION_FACTOR"`
	LoadFactor        float64 `mapstructure:"SHARDING_LOAD_FACTOR"`
}

// initShardingConfig initializes the sharding configuration from environment variables.
func initShardingConfig() *ShardingConfig {
	// Set defaults
	viper.SetDefault("SHARDING_SHARD_COUNT", pkgconstant.SSEShardCount)
	viper.SetDefault("SHARDING_REPLICATION_FACTOR", pkgconstant.DefaultReplicationFactor)
	viper.SetDefault("SHARDING_LOAD_FACTOR", pkgconstant.DefaultLoadFactor)

	shardingConfig := &ShardingConfig{}
	if err := viper.Unmarshal(shardingConfig); err != nil {
		panic(err)
	}

	return shardingConfig
}
