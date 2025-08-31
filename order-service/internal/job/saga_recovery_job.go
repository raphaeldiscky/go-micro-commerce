package job

import (
	"context"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/saga"
)

// SagaRecoveryJob handles recovery of failed sagas.
type SagaRecoveryJob struct {
	orchestrator saga.Orchestrator
	logger       logger.Logger
	interval     time.Duration
	enabled      bool
}

// NewSagaRecoveryJob creates a new saga recovery job.
func NewSagaRecoveryJob(
	orchestrator saga.Orchestrator,
	appLogger logger.Logger,
	interval time.Duration,
) *SagaRecoveryJob {
	return &SagaRecoveryJob{
		orchestrator: orchestrator,
		logger:       appLogger,
		interval:     interval,
		enabled:      true,
	}
}

// Start begins the recovery job.
func (j *SagaRecoveryJob) Start(ctx context.Context) {
	j.logger.Info("Starting saga recovery job")

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			j.logger.Info("Stopping saga recovery job")

			return
		case <-ticker.C:
			if j.enabled {
				j.recoverSagas(ctx)
			}
		}
	}
}

// recoverSagas attempts to recover failed sagas.
func (j *SagaRecoveryJob) recoverSagas(ctx context.Context) {
	j.logger.Debug("Running saga recovery check")

	recoveryCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := j.orchestrator.RecoverFailedSagas(recoveryCtx); err != nil {
		j.logger.Errorf("Saga recovery failed: %v", err)
	}
}

// Stop disables the recovery job.
func (j *SagaRecoveryJob) Stop() {
	j.enabled = false
}

// Resume enables the recovery job.
func (j *SagaRecoveryJob) Resume() {
	j.enabled = true
}
