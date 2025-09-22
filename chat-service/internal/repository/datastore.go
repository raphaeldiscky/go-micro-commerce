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
	ConversationRepository() ConversationRepository
	MessageRepository() MessageRepository
	ParticipantRepository() ParticipantRepository
	ConnectionRepository() ConnectionRepository
}

// dataStore is a struct that implements the DataStore interface.
type dataStore struct {
	pool *pgxpool.Pool
	db   DBTX
}

// NewDataStore creates a new DataStore.
func NewDataStore(pool *pgxpool.Pool, _ *redislock.Client, _ logger.Logger) DataStore {
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

// ConversationRepository returns a new ConversationRepository.
func (s *dataStore) ConversationRepository() ConversationRepository {
	return NewConversationRepository(s.db)
}

// MessageRepository returns a new MessageRepository.
func (s *dataStore) MessageRepository() MessageRepository {
	return NewMessageRepository(s.db)
}

// ParticipantRepository returns a new ParticipantRepository.
func (s *dataStore) ParticipantRepository() ParticipantRepository {
	return NewParticipantRepository(s.db)
}

// ConnectionRepository returns a new ConnectionRepository.
func (s *dataStore) ConnectionRepository() ConnectionRepository {
	return NewConnectionRepository(s.db)
}
