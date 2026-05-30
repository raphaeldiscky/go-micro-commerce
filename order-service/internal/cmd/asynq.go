package cmd

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/provider"
)

// asynqRunner wraps the asynq server as a Runner.
type asynqRunner struct {
	asynqProvider *provider.AsynqProvider
	logger        logger.Logger
}

// newAsynqRunner creates a new asynq runner.
func newAsynqRunner(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) (*asynqRunner, error) {
	asynqProvider, err := provider.SetupAsynq(cfg, providers, appLogger)
	if err != nil {
		return nil, err
	}

	return &asynqRunner{
		asynqProvider: asynqProvider,
		logger:        appLogger,
	}, nil
}

// Name returns the name of the runner.
func (r *asynqRunner) Name() string {
	return "Asynq Worker"
}

// Start starts the asynq worker.
func (r *asynqRunner) Start(ctx context.Context) error {
	r.logger.Info("Starting asynq worker...")

	errChan := make(chan error, 1)

	go func() {
		if err := r.asynqProvider.Server.Start(r.asynqProvider.Mux); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil // Context canceled, normal shutdown
	case err := <-errChan:
		return err // Server error
	}
}

// Shutdown gracefully shuts down the asynq worker.
func (r *asynqRunner) Shutdown(_ context.Context) error {
	r.logger.Info("Stopping asynq worker...")
	r.asynqProvider.Server.Stop()
	r.logger.Info("Asynq worker stopped successfully")

	return nil
}

// newAsynqCmd runs the asynq worker role.
func newAsynqCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "asynq",
		Short: "Run the asynq task worker",
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := bootstrap(cmd.Context())
			if err != nil {
				return err
			}
			defer app.stop()

			runner, err := newAsynqRunner(app.cfg, app.logger, app.providers)
			if err != nil {
				return err
			}

			return newManager(app.cfg, app.logger).run(app.ctx, runner)
		},
	}
}
