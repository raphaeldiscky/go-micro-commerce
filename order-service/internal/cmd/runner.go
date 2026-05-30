// Package cmd hosts the cobra subcommands for the order-service binary.
// Each subcommand maps to one deployable role (serve, kafka-consumer,
// outbox, scheduler, temporal, asynq); the "all" command runs every role
// in a single process for single-deployment setups.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/provider"
)

// Runner is a long-running process component with a managed lifecycle.
type Runner interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Name() string
}

// appContext holds the shared dependencies built once per role invocation.
type appContext struct {
	ctx       context.Context
	stop      context.CancelFunc
	cfg       *config.Config
	logger    logger.Logger
	providers *provider.Providers
}

// bootstrap performs the common per-role startup: a signal-aware context,
// configuration, logger, global providers and the asynq client. The
// returned stop func must be deferred by the caller on success.
func bootstrap(parent context.Context) (*appContext, error) {
	ctx, stop := signal.NotifyContext(parent,
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	cfg, err := config.LoadConfig()
	if err != nil {
		stop()

		return nil, fmt.Errorf("load config: %w", err)
	}

	appLogger := logger.NewLogrusLogger(cfg.App.LoggerLevel)

	providers, err := provider.SetupGlobal(ctx, cfg, appLogger)
	if err != nil {
		stop()

		return nil, fmt.Errorf("setup providers: %w", err)
	}

	// Asynq client/inspector must be resolved before runners that depend
	// on them (the order service enqueues asynq tasks) are constructed.
	if err = provider.SetupAsynqClient(cfg, providers, appLogger); err != nil {
		stop()

		return nil, fmt.Errorf("setup asynq client: %w", err)
	}

	return &appContext{
		ctx:       ctx,
		stop:      stop,
		cfg:       cfg,
		logger:    appLogger,
		providers: providers,
	}, nil
}

// roleCmd builds a cobra command for a single role. build returns the
// runners to manage and an optional cleanup func (e.g. Consul deregister)
// invoked on shutdown.
func roleCmd(use, short string, build func(*appContext) ([]Runner, func())) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := bootstrap(cmd.Context())
			if err != nil {
				return err
			}
			defer app.stop()

			runners, cleanup := build(app)
			if cleanup != nil {
				defer cleanup()
			}

			return newManager(app.cfg, app.logger).run(app.ctx, runners...)
		},
	}
}

// manager runs a set of runners and coordinates their graceful shutdown.
type manager struct {
	cfg     *config.Config
	logger  logger.Logger
	runners []Runner
	wg      sync.WaitGroup
}

// newManager creates a manager for the given config and logger.
func newManager(cfg *config.Config, appLogger logger.Logger) *manager {
	return &manager{cfg: cfg, logger: appLogger}
}

// run starts the given runners, blocks until the context is canceled, then
// performs graceful shutdown.
func (m *manager) run(ctx context.Context, runners ...Runner) error {
	m.runners = runners

	for _, r := range runners {
		m.wg.Add(1)

		runner := r // capture loop variable

		go func() {
			defer m.wg.Done()

			m.logger.Infof("Starting runner: %s", runner.Name())

			if err := runner.Start(ctx); err != nil {
				m.logger.Errorf("Runner %s failed: %v", runner.Name(), err)
			}
		}()
	}

	m.logger.Info("All runners started successfully")

	// Wait for context cancellation (shutdown signal)
	<-ctx.Done()
	m.logger.Info("Shutdown signal received, initiating graceful shutdown...")

	return m.shutdown()
}

// shutdown performs graceful shutdown of all runners in reverse order.
func (m *manager) shutdown() error {
	m.logger.Info("Starting graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), m.cfg.App.TimeoutShutdown)
	defer cancel()

	var shutdownErrors []error

	for i := len(m.runners) - 1; i >= 0; i-- {
		runner := m.runners[i]
		m.logger.Infof("Shutting down runner: %s", runner.Name())

		if err := runner.Shutdown(shutdownCtx); err != nil {
			m.logger.Errorf("Error shutting down runner %s: %v", runner.Name(), err)
			shutdownErrors = append(shutdownErrors, err)
		} else {
			m.logger.Infof("Runner %s shut down successfully", runner.Name())
		}
	}

	// Wait for all runners to finish with timeout
	done := make(chan struct{})

	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("All runners stopped gracefully")
	case <-shutdownCtx.Done():
		m.logger.Warn("Shutdown timeout reached, some runners may not have stopped gracefully")
	}

	if len(shutdownErrors) > 0 {
		return fmt.Errorf("shutdown completed with %d errors", len(shutdownErrors))
	}

	m.logger.Info("Graceful shutdown completed successfully")

	return nil
}
