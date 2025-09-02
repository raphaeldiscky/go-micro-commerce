package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/job"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/provider"
)

// JobSchedulerWorker wraps the job scheduler as a Worker.
type JobSchedulerWorker struct {
	scheduler *job.Scheduler
	logger    logger.Logger
}

// NewJobSchedulerWorker creates a new job scheduler worker.
func NewJobSchedulerWorker(
	appLogger logger.Logger,
	providers *provider.Providers,
) *JobSchedulerWorker {
	return &JobSchedulerWorker{
		scheduler: providers.JobScheduler,
		logger:    appLogger,
	}
}

// Start starts the job scheduler.
func (w *JobSchedulerWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting job scheduler worker")

	return w.scheduler.Start(ctx)
}

// Shutdown shuts down the job scheduler.
func (w *JobSchedulerWorker) Shutdown(ctx context.Context) error {
	w.logger.Info("Shutting down job scheduler worker")

	return w.scheduler.Shutdown(ctx)
}

// Name returns the name of the worker.
func (w *JobSchedulerWorker) Name() string {
	return "job-scheduler"
}
