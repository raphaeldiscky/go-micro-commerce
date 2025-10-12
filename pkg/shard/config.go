package shard

// Config holds configuration for the shard resolver.
type Config struct {
	ShardCount        int     `json:"shardCount"`
	ReplicationFactor int     `json:"replicationFactor"`
	LoadFactor        float64 `json:"loadFactor"`
}

const (
	defaultShardCount        = 10
	defaultReplicationFactor = 20
	defaultLoadFactor        = 1.25
)

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		ShardCount:        defaultShardCount,
		ReplicationFactor: defaultReplicationFactor,
		LoadFactor:        defaultLoadFactor,
	}
}
