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
	CartRepository() CartRepository
	CheckoutSessionRepository() CheckoutSessionRepository
	OutboxRepository() OutboxRepository
	LockRepository() LockRepository
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

	err = fn(&dataStore{
		pool:   s.pool,
		db:     tx,
		rdl:    s.rdl,
		logger: s.logger,
	})
	if err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			return err
		}

		return err
	}

	return tx.Commit(ctx)
}

// CartRepository returns a new CartRepository.
func (s *dataStore) CartRepository() CartRepository {
	return NewCartRepository(s.db, s.logger)
}

// CheckoutSessionRepository returns a new CheckoutSessionRepository.
func (s *dataStore) CheckoutSessionRepository() CheckoutSessionRepository {
	return NewCheckoutSessionRepository(s.db)
}

// OutboxRepository returns a new OutboxRepository.
func (s *dataStore) OutboxRepository() OutboxRepository {
	return NewOutboxRepository(s.db)
}

// LockRepository returns a new LockRepository.
func (s *dataStore) LockRepository() LockRepository {
	return NewLockRepository(s.rdl, s.logger)
}
