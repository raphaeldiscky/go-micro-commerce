// Package cmd hosts the cobra subcommands for the product-service binary.
// Each subcommand maps to one deployable role (serve, grpc); the "all"
// command runs every role in a single process for single-deployment setups.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/provider"
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
	telemetry *telemetry.Telemetry
	providers *provider.Providers
}

// bootstrap performs the common per-role startup: configuration, logger,
// telemetry, a signal-aware context and global providers. The returned stop
// func tears down telemetry and the signal context and must be deferred by
// the caller on success.
func bootstrap(parent context.Context) (*appContext, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	appLogger := logger.NewLogrusLogger(cfg.App.LoggerLevel)

	tel, telemetryCleanup := setupTelemetry(cfg, appLogger)

	ctx, stop := signal.NotifyContext(parent,
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	providers, err := provider.SetupGlobal(ctx, cfg, appLogger)
	if err != nil {
		stop()
		telemetryCleanup()

		return nil, fmt.Errorf("setup providers: %w", err)
	}

	return &appContext{
		ctx: ctx,
		stop: func() {
			stop()
			telemetryCleanup()
		},
		cfg:       cfg,
		logger:    appLogger,
		telemetry: tel,
		providers: providers,
	}, nil
}

// setupTelemetry initializes OpenTelemetry tracing and Prometheus metrics.
func setupTelemetry(cfg *config.Config, appLogger logger.Logger) (*telemetry.Telemetry, func()) {
	tel, err := telemetry.NewTelemetry(telemetry.Config{
		TracingEnabled:       cfg.Tracing.Enabled,
		TracingURL:           cfg.Tracing.URL,
		TracingServiceName:   cfg.Tracing.ServiceName,
		TracingSamplingRate:  cfg.Tracing.SamplingRate,
		TracingEnvironment:   cfg.Tracing.Environment,
		TracingBatchTimeout:  cfg.Tracing.BatchTimeout,
		TracingExportTimeout: cfg.Tracing.ExportTimeout,
		MetricsEnabled:       cfg.Metrics.Enabled,
		MetricsPath:          cfg.Metrics.Path,
	})
	if err != nil {
		appLogger.Errorf("Failed to initialize telemetry: %v", err)

		return nil, func() {}
	}

	appLogger.Info("Telemetry initialized successfully")

	return tel, func() {
		ctx, cancel := context.WithTimeout(context.Background(), constant.MetricTimeout)
		defer cancel()

		if err = tel.Shutdown(ctx); err != nil {
			appLogger.Errorf("Failed to shutdown telemetry: %v", err)
		} else {
			appLogger.Info("Telemetry shutdown successfully")
		}
	}
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
