package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/job"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/saga"
)

// SetupJobScheduler creates and configures the job scheduler with all jobs.
func SetupJobScheduler(
	cfg *config.Config,
	sagaOrchestrator saga.Orchestrator,
	appLogger logger.Logger,
	providers *Providers,
) *job.Scheduler {
	scheduler := job.NewScheduler(appLogger)

	// Register saga recovery job if enabled
	if cfg.Jobs.SagaRecovery.Enabled {
		sagaRecoveryJob := job.NewSagaRecoveryJob(
			sagaOrchestrator,
			providers.DataStore,
			appLogger,
			cfg.Jobs.SagaRecovery.Interval,
		)

		// Apply configuration overrides
		sagaRecoveryJob.SetMaxRetries(cfg.Jobs.SagaRecovery.MaxRetries)
		sagaRecoveryJob.SetMaxAge(cfg.Jobs.SagaRecovery.MaxAge)

		scheduler.RegisterJob(sagaRecoveryJob)
	}

	// Register additional jobs here as they are implemented
	// if cfg.Jobs.Cleanup != nil && cfg.Jobs.Cleanup.Enabled {
	//     cleanupJob := job.NewCleanupJob(...)
	//     scheduler.RegisterJob(cleanupJob)
	// }

	return scheduler
}
