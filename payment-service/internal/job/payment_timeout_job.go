// Package job provides background jobs for the payment service.
package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/service"
)

// PaymentTimeoutJob handles automatic timeout of expired payments after 24-hour window.
type PaymentTimeoutJob struct {
	paymentService service.PaymentService
	dataStore      repository.DataStore
	logger         logger.Logger
	config         *config.Config
	interval       time.Duration
	enabled        bool
	running        bool
	mu             sync.RWMutex
	batchSize      int
}

// NewPaymentTimeoutJob creates a new payment timeout job.
func NewPaymentTimeoutJob(
	paymentService service.PaymentService,
	dataStore repository.DataStore,
	config *config.Config,
	appLogger logger.Logger,
	interval time.Duration,
) *PaymentTimeoutJob {
	return &PaymentTimeoutJob{
		paymentService: paymentService,
		dataStore:      dataStore,
		logger:         appLogger,
		config:         config,
		interval:       interval,
		enabled:        true,
		batchSize:      config.Job.PaymentTimeout.BatchSize,
	}
}

// Name returns the name of the job.
func (j *PaymentTimeoutJob) Name() string {
	return "payment-timeout"
}

// Interval returns the execution interval of the job.
func (j *PaymentTimeoutJob) Interval() time.Duration {
	return j.interval
}

// IsEnabled returns whether the job is enabled.
func (j *PaymentTimeoutJob) IsEnabled() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.enabled
}

// Start executes a single iteration of the payment timeout job.
func (j *PaymentTimeoutJob) Start(ctx context.Context) {
	if !j.IsEnabled() {
		return
	}

	j.logger.Debug("Executing payment timeout job")
	j.timeoutExpiredPayments(ctx)
}

// timeoutExpiredPayments finds and times out all expired pending payments.
func (j *PaymentTimeoutJob) timeoutExpiredPayments(ctx context.Context) {
	j.mu.RLock()

	if j.running {
		j.mu.RUnlock()
		j.logger.Debug("Payment timeout job already running, skipping")

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

	// Acquire distributed lock to prevent concurrent execution across multiple instances
	lockRepo := j.dataStore.LockRepository()
	lockKey := "payment:timeout:lock"

	lock, err := lockRepo.Get(
		ctx,
		lockKey,
		j.config.Job.PaymentTimeout.RedisLockTTL,
		&redislock.Options{
			RetryStrategy: redislock.LimitRetry(
				redislock.LinearBackoff(j.config.Job.PaymentTimeout.RedisLockBackoff),
				j.config.Job.PaymentTimeout.RedisLockMaxRetries,
			),
		},
	)
	if err != nil {
		j.logger.Debugf("Could not acquire timeout lock, another instance may be running: %v", err)

		return
	}

	defer func() {
		if err = lockRepo.Release(ctx, lock); err != nil {
			j.logger.Warnf("Failed to release timeout lock: %v", err)
		}
	}()

	timeoutCtx, cancel := context.WithTimeout(ctx, j.config.Job.PaymentTimeout.Timeout)
	defer cancel()

	if err = j.processExpiredPayments(timeoutCtx); err != nil {
		j.logger.Errorf("Payment timeout processing failed: %v", err)
	}
}

// processExpiredPayments retrieves and times out expired payments in batches.
func (j *PaymentTimeoutJob) processExpiredPayments(ctx context.Context) error {
	paymentRepo := j.dataStore.PaymentRepository()

	// Find expired payments that are still pending
	expiredPayments, err := paymentRepo.FindExpiredPayments(ctx, j.batchSize)
	if err != nil {
		return fmt.Errorf("failed to find expired payments: %w", err)
	}

	if len(expiredPayments) == 0 {
		j.logger.Debug("No expired payments found to timeout")

		return nil
	}

	j.logger.Infof("Found %d expired payments to timeout", len(expiredPayments))

	successCount := 0
	failureCount := 0

	for _, payment := range expiredPayments {
		if err = ctx.Err(); err != nil {
			j.logger.Warn("Context cancelled, stopping payment timeout processing")

			break
		}

		j.logger.Infof(
			"Timing out payment %s for order %s (expired at: %v)",
			payment.ID,
			payment.OrderID,
			payment.ExpiresAt,
		)

		// Call payment service to timeout the payment
		// This will update status and publish PaymentTimeoutEvent
		if err = j.paymentService.TimeoutPayment(ctx, payment.OrderID); err != nil {
			j.logger.Errorf(
				"Failed to timeout payment %s for order %s: %v",
				payment.ID,
				payment.OrderID,
				err,
			)

			failureCount++

			continue
		}

		successCount++

		j.logger.Infof(
			"Successfully timed out payment %s for order %s",
			payment.ID,
			payment.OrderID,
		)
	}

	j.logger.Infof(
		"Payment timeout job completed: %d succeeded, %d failed, %d total",
		successCount,
		failureCount,
		len(expiredPayments),
	)

	return nil
}

// Stop disables the payment timeout job.
func (j *PaymentTimeoutJob) Stop() {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.enabled = false
	j.logger.Info("Payment timeout job stopped")
}

// Resume enables the payment timeout job.
func (j *PaymentTimeoutJob) Resume() {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.enabled = true
	j.logger.Info("Payment timeout job resumed")
}

// IsRunning returns true if the job is currently running.
func (j *PaymentTimeoutJob) IsRunning() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()

	return j.running
}

// SetBatchSize sets the number of payments to process per execution.
func (j *PaymentTimeoutJob) SetBatchSize(batchSize int) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.batchSize = batchSize
	j.logger.Infof("Payment timeout job batch size set to %d", batchSize)
}
