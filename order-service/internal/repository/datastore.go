package repository

import (
	"context"

	"github.com/bsm/redislock"
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
	OrderRepository() OrderRepositoryInterface
	ProductRepository() ProductRepositoryInterface
	LockRepository() LockRepositoryInterface
	OutboxRepository() OutboxRepositoryInterface
	SagaStateRepository() SagaStateRepositoryInterface
}

// dataStore is a struct that implements the DataStore interface.
type dataStore struct {
	pool *pgxpool.Pool
	db   DBTX
	rdl  *redislock.Client
}

// NewDataStore creates a new DataStore.
func NewDataStore(pool *pgxpool.Pool, rdl *redislock.Client) DataStore {
	return &dataStore{
		pool: pool,
		db:   pool,
		rdl:  rdl,
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

// OrderRepository returns a new OrderRepository.
func (s *dataStore) OrderRepository() OrderRepositoryInterface {
	return NewOrderRepositoryPostgres(s.db)
}

// ProductRepository returns a new ProductRepository.
func (s *dataStore) ProductRepository() ProductRepositoryInterface {
	return NewProductRepositoryPostgres(s.db)
}

// LockRepository returns a new LockRepository.
func (s *dataStore) LockRepository() LockRepositoryInterface {
	return NewLockRepository(s.rdl)
}

// OutboxRepository returns a new OutboxRepository.
func (s *dataStore) OutboxRepository() OutboxRepositoryInterface {
	return NewOutboxRepository(s.db)
}

// SagaStateRepository returns a new SagaStateRepository.
func (s *dataStore) SagaStateRepository() SagaStateRepositoryInterface {
	return NewSagaStateRepository(s.db)
}
