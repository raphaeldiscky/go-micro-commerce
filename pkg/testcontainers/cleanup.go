package testcontainers

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	redisclient "github.com/redis/go-redis/v9"
)

// TruncateTables truncates the specified tables in PostgreSQL.
// Uses CASCADE to handle foreign key constraints and RESTART IDENTITY to reset sequences.
func TruncateTables(ctx context.Context, pool *pgxpool.Pool, tables []string) error {
	if len(tables) == 0 {
		return nil
	}

	// Build TRUNCATE statement with all tables
	// TRUNCATE is more efficient than DELETE for clearing entire tables
	tableList := strings.Join(tables, ", ")
	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableList)

	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	return nil
}

// FlushRedis flushes all data from Redis.
// Use with caution - this clears ALL databases in the Redis instance.
func FlushRedis(ctx context.Context, client *redisclient.Client) error {
	if err := client.FlushAll(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush redis: %w", err)
	}

	return nil
}
