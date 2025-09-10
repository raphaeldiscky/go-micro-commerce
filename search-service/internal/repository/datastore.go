// Package repository provides data access layer implementations for the notification service.
package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/client"
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
	InboxRepository() InboxRepositoryInterface
	SearchRepository() SearchRepositoryInterface
}

// dataStore is a struct that implements the DataStore interface.
type dataStore struct {
	pool          *pgxpool.Pool
	db            DBTX
	elasticClient client.ElasticsearchClientInterface
	logger        logger.Logger
}

// NewDataStore creates a new DataStore.
func NewDataStore(
	pool *pgxpool.Pool,
	elasticClient client.ElasticsearchClientInterface,
	appLogger logger.Logger,
) DataStore {
	return &dataStore{
		pool:          pool,
		db:            pool,
		elasticClient: elasticClient,
		logger:        appLogger,
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

// InboxRepository returns a new InboxRepository.
func (s *dataStore) InboxRepository() InboxRepositoryInterface {
	return NewInboxRepository(s.db)
}

// SearchRepository returns a new SearchRepository.
func (s *dataStore) SearchRepository() SearchRepositoryInterface {
	return NewSearchRepository(s.elasticClient, s.logger)
}
