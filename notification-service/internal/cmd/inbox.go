package cmd

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/worker"
)

// inboxProcessorRunner wraps the inbox processor as a Runner.
type inboxProcessorRunner struct {
	consumer *worker.InboxProcessor
}

// newInboxProcessorRunner creates a new inbox processor runner.
func newInboxProcessorRunner(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) *inboxProcessorRunner {
	return &inboxProcessorRunner{
		consumer: provider.SetupInboxProcessor(cfg, appLogger, providers),
	}
}

// Name returns the name of the runner.
func (r *inboxProcessorRunner) Name() string {
	return "Inbox Processor"
}

// Start starts the inbox processor.
func (r *inboxProcessorRunner) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if err := r.consumer.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil // Context canceled, normal shutdown
	case err := <-errChan:
		return err // Consumer error
	}
}

// Shutdown gracefully shuts down the inbox processor.
func (r *inboxProcessorRunner) Shutdown(ctx context.Context) error {
	return r.consumer.Shutdown(ctx)
}

// newInboxCmd runs the inbox processor role.
func newInboxCmd() *cobra.Command {
	return roleCmd(
		"inbox",
		"Run the inbox processor",
		func(app *appContext) ([]Runner, func()) {
			runner := newInboxProcessorRunner(app.cfg, app.logger, app.providers)

			return []Runner{runner}, nil
		},
	)
}
