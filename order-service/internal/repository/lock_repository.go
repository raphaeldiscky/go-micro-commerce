package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bsm/redislock"
	"github.com/raphaeldiscky/go-micro-template/pkg/httperror"
)

// LockRepositoryInterface defines the interface for acquiring distributed locks.
type LockRepositoryInterface interface {
	Get(
		ctx context.Context,
		key string,
		ttl time.Duration,
		opt *redislock.Options,
	) (*redislock.Lock, error)
	Release(ctx context.Context, lock *redislock.Lock) error
}

// LockRepository is the implementation of LockRepositoryInterface.
type LockRepository struct {
	rdl *redislock.Client
}

// NewLockRepository creates a new instance of LockRepository.
func NewLockRepository(rdl *redislock.Client) LockRepositoryInterface {
	return &LockRepository{
		rdl: rdl,
	}
}

// Get acquires a distributed lock.
func (r *LockRepository) Get(
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

		log.Printf("failed to obtain lock: %v", err)

		return nil, err
	}

	return lock, nil
}

// Release releases a distributed lock.
func (r *LockRepository) Release(ctx context.Context, lock *redislock.Lock) error {
	if err := lock.Release(ctx); err != nil {
		log.Printf("failed to release lock: %v", err)

		return err
	}

	return nil
}
