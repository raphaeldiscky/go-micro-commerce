// Package worker provides the entry point for starting the worker services.
package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/provider"
)

// Start initializes and starts the worker services.
func Start(cfg *config.Config, appLogger logger.Logger) {
	providers, err := provider.SetupGlobal(cfg)
	if err != nil {
		appLogger.Fatal("failed to setup providers:", err)
	}

	// Initialize ProductService early to avoid race conditions
	provider.InitializeProductService(cfg, appLogger, providers)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rootCmd := &cobra.Command{
		Use: "product-service",
	}
	cmd := []*cobra.Command{
		{
			Use:   "serve-all",
			Short: "Run all",
			Run: func(_ *cobra.Command, _ []string) {
				go runHTTPWorker(ctx, cfg, appLogger, providers)
				go runGRPCWorker(ctx, cfg, appLogger, providers)
				go runKafkaConsumerWorker(ctx, cfg, appLogger, providers)

				<-ctx.Done()
			},
		},
	}

	rootCmd.AddCommand(cmd...)

	if err := rootCmd.Execute(); err != nil {
		appLogger.Fatal(err)
	}
}
