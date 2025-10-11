// Package repository provides data access layer implementations for the notification service.
package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	InboxRepository() InboxRepository
	NotificationRepository() NotificationRepository
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

// InboxRepository returns a new InboxRepository.
func (s *dataStore) InboxRepository() InboxRepository {
	return NewInboxRepository(s.db)
}

// NotificationRepository returns a new NotificationRepository.
func (s *dataStore) NotificationRepository() NotificationRepository {
	return NewNotificationRepository(s.db)
}
