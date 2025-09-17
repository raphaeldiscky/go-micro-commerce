package repository

import (
	"context"
	"errors"
	"time"

	"github.com/bsm/redislock"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// LockRepository defines the interface for acquiring distributed locks.
type LockRepository interface {
	Get(
		ctx context.Context,
		key string,
		ttl time.Duration,
		opt *redislock.Options,
	) (*redislock.Lock, error)
	Release(ctx context.Context, lock *redislock.Lock) error
}

// lockRepository is the implementation of LockRepository.
type lockRepository struct {
	rdl    *redislock.Client
	logger logger.Logger
}

// NewLockRepository creates a new instance of lockRepository.
func NewLockRepository(rdl *redislock.Client, appLogger logger.Logger) LockRepository {
	return &lockRepository{
		rdl:    rdl,
		logger: appLogger,
	}
}

// Get acquires a distributed lock.
func (r *lockRepository) Get(
	ctx context.Context,
	key string,
	ttl time.Duration,
	opt *redislock.Options,
) (*redislock.Lock, error) {
	lock, err := r.rdl.Obtain(ctx, key, ttl, opt)
	if err != nil {
		if errors.Is(err, redislock.ErrNotObtained) {
			return nil, httperror.NewRequestDuplicateError()
		}

		r.logger.Printf("failed to obtain lock: %v", err)

		return nil, err
	}

	return lock, nil
}

// Release releases a distributed lock.
func (r *lockRepository) Release(ctx context.Context, lock *redislock.Lock) error {
	if err := lock.Release(ctx); err != nil {
		r.logger.Printf("failed to release lock: %v", err)

		return err
	}

	return nil
}
