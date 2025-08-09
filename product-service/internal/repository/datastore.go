package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type DataStore interface {
	WithTransaction(ctx context.Context, fn func(DataStore) error) error
	ProductRepository() ProductRepository
}

type dataStore struct {
	pool *pgxpool.Pool
	db   DBTX
}

func NewDataStore(pool *pgxpool.Pool) DataStore {
	return &dataStore{
		pool: pool,
		db:   pool,
	}
}

func (s *dataStore) WithTransaction(ctx context.Context, fn func(DataStore) error) error {
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

func (s *dataStore) ProductRepository() ProductRepository {
	return NewProductRepositoryPostgres(s.db)
}
