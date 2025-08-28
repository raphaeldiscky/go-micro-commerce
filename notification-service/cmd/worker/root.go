// Package worker provides the entry point for starting the worker services.
package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/provider"
)

// Start initializes the application workers.
func Start(cfg *config.Config, appLogger logger.Logger) {
	_, err := provider.SetupGlobal(cfg)
	if err != nil {
		appLogger.Fatal("Failed to setup providers:", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rootCmd := &cobra.Command{
		Use: "notification-service",
	}
	cmd := []*cobra.Command{
		{
			Use:   "serve-all",
			Short: "Run all",
			Run: func(_ *cobra.Command, _ []string) {
				go runHTTPWorker(ctx, cfg, appLogger)
				go runKafkaConsumerWorker(ctx, cfg, appLogger)

				<-ctx.Done()
			},
		},
	}

	rootCmd.AddCommand(cmd...)

	if err := rootCmd.Execute(); err != nil {
		appLogger.Fatal(err)
	}
}
