// Package job provides background jobs for the order service.
package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/saga"
)

// SagaRecoveryJob handles recovery of failed sagas.
type SagaRecoveryJob struct {
	orchestrator saga.Orchestrator
	dataStore    repository.DataStore
	logger       logger.Logger
	interval     time.Duration
	enabled      bool
	running      bool
	mu           sync.RWMutex
	maxRetries   int64
	maxAge       time.Duration
}

// NewSagaRecoveryJob creates a new saga recovery job.
func NewSagaRecoveryJob(
	orchestrator saga.Orchestrator,
	dataStore repository.DataStore,
	appLogger logger.Logger,
	interval time.Duration,
) *SagaRecoveryJob {
	return &SagaRecoveryJob{
		orchestrator: orchestrator,
		dataStore:    dataStore,
		logger:       appLogger,
		interval:     interval,
		enabled:      true,
		maxRetries:   5,
		maxAge:       24 * time.Hour,
	}
}

// Name returns the name of the job.
func (j *SagaRecoveryJob) Name() string {
	return "saga-recovery"
}

// Interval returns the execution interval of the job.
func (j *SagaRecoveryJob) Interval() time.Duration {
	return j.interval
}

// IsEnabled returns whether the job is enabled.
func (j *SagaRecoveryJob) IsEnabled() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.enabled
}

// Start executes a single iteration of the recovery job.
func (j *SagaRecoveryJob) Start(ctx context.Context) {
	if !j.IsEnabled() {
		return
	}

	j.logger.Debug("Executing saga recovery job")
	j.recoverSagas(ctx)
}

// recoverSagas attempts to recover failed sagas with distributed locking.
func (j *SagaRecoveryJob) recoverSagas(ctx context.Context) {
	j.mu.RLock()

	if j.running {
		j.mu.RUnlock()
		j.logger.Debug("Saga recovery already running, skipping")

		return
	}

	j.mu.RUnlock()

	j.mu.Lock()
	j.running = true
	j.mu.Unlock()

	defer func() {
		j.mu.Lock()
		j.running = false
		j.mu.Unlock()
	}()

	j.logger.Debug("Running saga recovery check")

	// Acquire distributed lock to prevent concurrent recovery
	lockRepo := j.dataStore.LockRepository()
	lockKey := "saga:recovery:lock"

	lock, err := lockRepo.Get(ctx, lockKey, 10*time.Minute, &redislock.Options{
		RetryStrategy: redislock.LimitRetry(
			redislock.LinearBackoff(100*time.Millisecond),
			3,
		),
	})
	if err != nil {
		j.logger.Debugf("Could not acquire recovery lock, another instance may be running: %v", err)

		return
	}

	defer func() {
		if err := lockRepo.Release(ctx, lock); err != nil {
			j.logger.Warnf("Failed to release recovery lock: %v", err)
		}
	}()

	recoveryCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := j.recoverFailedSagasWithRetries(recoveryCtx); err != nil {
		j.logger.Errorf("Saga recovery failed: %v", err)
	}

	if err := j.recoverTimeoutSagas(recoveryCtx); err != nil {
		j.logger.Errorf("Timeout saga recovery failed: %v", err)
	}
}

// recoverFailedSagasWithRetries recovers failed sagas with retry limits.
func (j *SagaRecoveryJob) recoverFailedSagasWithRetries(ctx context.Context) error {
	stateRepo := j.dataStore.SagaStateRepository()

	// Find failed sagas that can be retried
	failedSagas, err := stateRepo.FindPendingOrFailed(ctx, 50)
	if err != nil {
		return fmt.Errorf("failed to find sagas for recovery: %w", err)
	}

	if len(failedSagas) == 0 {
		j.logger.Debug("No failed sagas found for recovery")

		return nil
	}

	j.logger.Infof("Found %d failed sagas to recover", len(failedSagas))

	for _, sagaState := range failedSagas {
		// Check if saga can be retried
		if !sagaState.CanRetry(j.maxRetries, j.maxAge) {
			j.logger.Warnf(
				"Saga %s exceeded retry limits, marking as permanently failed",
				sagaState.ID,
			)
			// Mark as permanently failed
			if err := stateRepo.MarkAsFailed(ctx, sagaState.ID, "exceeded retry limits"); err != nil {
				j.logger.Errorf("Failed to mark saga as permanently failed: %v", err)
			}

			continue
		}

		j.logger.Infof("Attempting to recover saga %s for order %s (attempt %d)",
			sagaState.ID, sagaState.OrderID, sagaState.RetryCount+1)

		// Increment retry count
		sagaState.IncrementRetry()

		if err := stateRepo.UpdateWithVersion(ctx, sagaState); err != nil {
			j.logger.Errorf("Failed to update saga retry count: %v", err)

			continue
		}

		// Trigger recovery asynchronously
		go j.recoverSingleSaga(ctx, sagaState.OrderID)
	}

	return nil
}

// recoverTimeoutSagas handles sagas that have timed out.
func (j *SagaRecoveryJob) recoverTimeoutSagas(ctx context.Context) error {
	stateRepo := j.dataStore.SagaStateRepository()

	timeoutSagas, err := stateRepo.FindTimeoutSagas(ctx, 20)
	if err != nil {
		return fmt.Errorf("failed to find timeout sagas: %w", err)
	}

	if len(timeoutSagas) == 0 {
		return nil
	}

	j.logger.Infof("Found %d timeout sagas to handle", len(timeoutSagas))

	for _, sagaState := range timeoutSagas {
		j.logger.Warnf("Saga %s for order %s has timed out, starting compensation",
			sagaState.ID, sagaState.OrderID)

		// Mark as compensating and trigger compensation
		if err := stateRepo.MarkAsCompensating(ctx, sagaState.ID); err != nil {
			j.logger.Errorf("Failed to mark timeout saga as compensating: %v", err)

			continue
		}

		// Trigger compensation asynchronously
		go j.recoverSingleSaga(ctx, sagaState.OrderID)
	}

	return nil
}

// recoverSingleSaga recovers a single saga.
func (j *SagaRecoveryJob) recoverSingleSaga(_ context.Context, orderID uuid.UUID) {
	recoveryCtx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Minute,
	)
	defer cancel()

	if err := j.orchestrator.RecoverFailedSagas(recoveryCtx); err != nil {
		j.logger.Errorf("Failed to recover saga for order %s: %v", orderID, err)
	} else {
		j.logger.Infof("Successfully recovered saga for order %s", orderID)
	}
}

// Stop disables the recovery job.
func (j *SagaRecoveryJob) Stop() {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.enabled = false
	j.logger.Info("Saga recovery job stopped")
}

// Resume enables the recovery job.
func (j *SagaRecoveryJob) Resume() {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.enabled = true
}

// IsRunning returns true if recovery is currently running.
func (j *SagaRecoveryJob) IsRunning() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.running
}

// SetMaxRetries sets the maximum retry count for saga recovery.
func (j *SagaRecoveryJob) SetMaxRetries(maxRetries int64) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.maxRetries = maxRetries
}

// SetMaxAge sets the maximum age for saga recovery.
func (j *SagaRecoveryJob) SetMaxAge(maxAge time.Duration) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.maxAge = maxAge
}
