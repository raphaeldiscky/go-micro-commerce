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
	scheduler := job.NewScheduler(appLogger, cfg.Job)

	// Register saga recovery job if enabled
	if cfg.Job.Recovery.Enabled {
		sagaRecoveryJob := job.NewSagaRecoveryJob(
			sagaOrchestrator,
			providers.DataStore,
			cfg,
			appLogger,
			cfg.Job.Recovery.Interval,
		)

		// Apply configuration overrides
		sagaRecoveryJob.SetMaxRetries(cfg.Job.Recovery.MaxRetries)
		sagaRecoveryJob.SetMaxAge(cfg.Job.Recovery.MaxAge)

		scheduler.RegisterJob(sagaRecoveryJob)
	}

	// Register additional jobs here as they are implemented
	// if cfg.Job.Cleanup != nil && cfg.Job.Cleanup.Enabled {
	//     cleanupJob := job.NewCleanupJob(...)
	//     scheduler.RegisterJob(cleanupJob)
	// }

	return scheduler
}
