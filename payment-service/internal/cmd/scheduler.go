package cmd

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/job"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/provider"
)

// jobSchedulerRunner wraps the job scheduler as a Runner.
type jobSchedulerRunner struct {
	scheduler *job.Scheduler
	logger    logger.Logger
}

// newJobSchedulerRunner creates a new job scheduler runner.
func newJobSchedulerRunner(
	appLogger logger.Logger,
	providers *provider.Providers,
) *jobSchedulerRunner {
	return &jobSchedulerRunner{
		scheduler: providers.JobScheduler,
		logger:    appLogger,
	}
}

// Name returns the name of the runner.
func (r *jobSchedulerRunner) Name() string {
	return "job-scheduler"
}

// Start starts the job scheduler.
func (r *jobSchedulerRunner) Start(ctx context.Context) error {
	r.logger.Info("Starting job scheduler runner")

	return r.scheduler.Start(ctx)
}

// Shutdown shuts down the job scheduler.
func (r *jobSchedulerRunner) Shutdown(ctx context.Context) error {
	r.logger.Info("Shutting down job scheduler runner")

	return r.scheduler.Shutdown(ctx)
}

// newSchedulerCmd runs the job scheduler role.
func newSchedulerCmd() *cobra.Command {
	return roleCmd("scheduler", "Run the job scheduler", func(app *appContext) ([]Runner, func()) {
		runner := newJobSchedulerRunner(app.logger, app.providers)

		return []Runner{runner}, nil
	})
}
