package repository

import (
	"context"

	"github.com/bsm/redislock"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// DBTX is an interface that wraps the database transaction methods.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// DataStore is an interface that wraps the database access methods.
type DataStore interface {
	Atomic(ctx context.Context, fn func(DataStore) error) error
	OrderRepository() OrderRepositoryInterface
	LockRepository() LockRepositoryInterface
	OutboxRepository() OutboxRepositoryInterface
	SagaStateRepository() SagaStateRepositoryInterface
	InboxRepository() InboxRepositoryInterface
}

// dataStore is a struct that implements the DataStore interface.
type dataStore struct {
	pool   *pgxpool.Pool
	db     DBTX
	rdl    *redislock.Client
	logger logger.Logger
}

// NewDataStore creates a new DataStore.
func NewDataStore(pool *pgxpool.Pool, rdl *redislock.Client, appLogger logger.Logger) DataStore {
	return &dataStore{
		pool:   pool,
		db:     pool,
		rdl:    rdl,
		logger: appLogger,
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

// LockRepository returns a new LockRepository.
func (s *dataStore) LockRepository() LockRepositoryInterface {
	return NewLockRepository(s.rdl, s.logger)
}

// OutboxRepository returns a new OutboxRepository.
func (s *dataStore) OutboxRepository() OutboxRepositoryInterface {
	return NewOutboxRepository(s.db)
}

// SagaStateRepository returns a new SagaStateRepository.
func (s *dataStore) SagaStateRepository() SagaStateRepositoryInterface {
	return NewSagaStateRepository(s.db)
}

// InboxRepository returns a new InboxRepository.
func (s *dataStore) InboxRepository() InboxRepositoryInterface {
	return NewInboxRepository(s.db)
}
