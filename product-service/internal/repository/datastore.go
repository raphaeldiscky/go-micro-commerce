package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is an interface that wraps the database transaction methods.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// DataStore is an interface that wraps the database access methods.
type DataStore interface {
	Atomic(ctx context.Context, fn func(DataStore) error) error
	ProductRepository() ProductRepositoryInterface
}

// dataStore is a struct that implements the DataStore interface.
type dataStore struct {
	pool *pgxpool.Pool
	db   DBTX
}

// NewDataStore creates a new DataStore.
func NewDataStore(pool *pgxpool.Pool) DataStore {
	return &dataStore{
		pool: pool,
		db:   pool,
	}
}

// Atomic executes a function within a database transaction.
func (s *dataStore) Atomic(ctx context.Context, fn func(DataStore) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	err = fn(&dataStore{pool: s.pool, db: tx})
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
