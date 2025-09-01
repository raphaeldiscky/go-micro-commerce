package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// DBTX is an interface that wraps the database transaction methods.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

// DataStore is an interface that wraps the database access methods.
type DataStore interface {
	Atomic(ctx context.Context, fn func(DataStore) error) error
	ProductRepository() ProductRepositoryInterface
	CacheRepository() CacheRepositoryInterface
}

// dataStore is a struct that implements the DataStore interface.
type dataStore struct {
	pool        *pgxpool.Pool
	db          DBTX
	cacheClient redis.UniversalClient
}

// NewDataStore creates a new DataStore.
func NewDataStore(pool *pgxpool.Pool, cacheClient redis.UniversalClient) DataStore {
	return &dataStore{
		pool:        pool,
		db:          pool,
		cacheClient: cacheClient,
	}
}

// Atomic executes a function within a database transaction.
func (s *dataStore) Atomic(ctx context.Context, fn func(DataStore) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	err = fn(&dataStore{pool: s.pool, db: tx, cacheClient: s.cacheClient})
	if err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return err
		}

		return err
	}

	return tx.Commit(ctx)
}

// ProductRepository returns a new ProductRepository.
func (s *dataStore) ProductRepository() ProductRepositoryInterface {
	return NewProductRepositoryPostgres(s.db)
}

// CacheRepository returns a new CacheRepository.
func (s *dataStore) CacheRepository() CacheRepositoryInterface {
	return NewCacheRepositoryRedis(s.cacheClient)
}
