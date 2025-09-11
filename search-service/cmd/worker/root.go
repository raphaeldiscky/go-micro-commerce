// Package worker provides the entry point for starting the worker services.
package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/provider"
)

// Manager manages all workers and their lifecycle.
type Manager struct {
	cfg       *config.Config
	logger    logger.Logger
	providers *provider.Providers
	workers   []Worker
	wg        sync.WaitGroup
}

// Worker interface for all worker implementations.
type Worker interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Name() string
}

// Start initializes and starts the worker services.
func Start(ctx context.Context, cfg *config.Config, appLogger logger.Logger) error {
	providers, err := provider.SetupGlobal(ctx, cfg, appLogger)
	if err != nil {
		appLogger.Fatal("failed to setup providers:", err)
	}

	manager := &Manager{
		cfg:       cfg,
		logger:    appLogger,
		providers: providers,
	}

	rootCmd := &cobra.Command{
		Use: "search-service",
	}
	cmd := []*cobra.Command{
		{
			Use:   "serve-all",
			Short: "Run all workers",
			RunE: func(_ *cobra.Command, _ []string) error {
				return manager.runAllWorkers(ctx)
			},
		},
	}

	rootCmd.AddCommand(cmd...)

	return rootCmd.ExecuteContext(ctx)
}

func (wm *Manager) runAllWorkers(ctx context.Context) error {
	wm.logger.Info("Starting all workers...")

	// Initialize all workers
	workers := []Worker{
		NewHTTPWorker(wm.cfg, wm.logger, wm.providers),
		NewKafkaConsumerWorker(wm.cfg, wm.logger, wm.providers),
		NewInboxProcessorWorker(wm.cfg, wm.logger, wm.providers),
	}

	return wm.runWorkers(ctx, workers)
}

func (wm *Manager) runWorkers(ctx context.Context, workers []Worker) error {
	wm.workers = workers

	// Start all workers
	for _, w := range workers {
		wm.wg.Add(1)

		worker := w // capture loop variable

		go func() {
			defer wm.wg.Done()

			wm.logger.Infof("Starting worker: %s", worker.Name())

			if err := worker.Start(ctx); err != nil {
				wm.logger.Errorf("Worker %s failed: %v", worker.Name(), err)
			}
		}()
	}

	wm.logger.Info("All workers started successfully")

	// Wait for context cancellation (shutdown signal)
	<-ctx.Done()
	wm.logger.Info("Shutdown signal received, initiating graceful shutdown...")

	// Perform graceful shutdown
	return wm.shutdown()
}

// shutdown performs graceful shutdown of all workers.
func (wm *Manager) shutdown() error {
	wm.logger.Info("Starting graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), wm.cfg.App.TimeoutShutdown)
	defer cancel()

	var shutdownErrors []error

	for i := len(wm.workers) - 1; i >= 0; i-- {
		worker := wm.workers[i]
		wm.logger.Infof("Shutting down worker: %s", worker.Name())

		if err := worker.Shutdown(shutdownCtx); err != nil {
			wm.logger.Errorf("Error shutting down worker %s: %v", worker.Name(), err)
			shutdownErrors = append(shutdownErrors, err)
		} else {
			wm.logger.Infof("Worker %s shut down successfully", worker.Name())
		}
	}

	// Wait for all workers to finish with timeout
	done := make(chan struct{})

	go func() {
		wm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		wm.logger.Info("All workers stopped gracefully")
	case <-shutdownCtx.Done():
		wm.logger.Warn("Shutdown timeout reached, some workers may not have stopped gracefully")
	}

	if len(shutdownErrors) > 0 {
		return fmt.Errorf("shutdown completed with %d errors", len(shutdownErrors))
	}

	wm.logger.Info("Graceful shutdown completed successfully")

	return nil
}
