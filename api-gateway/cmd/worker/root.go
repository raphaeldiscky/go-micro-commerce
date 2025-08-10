// Package worker provides the entry point for starting the worker services.
package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/spf13/cobra"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/gateway"
)

// Start initializes the application workers.
func Start(cfg *config.Config, appLogger logger.Logger, gw *gateway.Gateway) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rootCmd := &cobra.Command{}
	cmd := []*cobra.Command{
		{
			Use:   "serve-all",
			Short: "Run all",
			Run: func(_ *cobra.Command, _ []string) {
				runHTTPWorker(ctx, cfg, appLogger, gw)
			},
		},
	}

	rootCmd.AddCommand(cmd...)

	if err := rootCmd.Execute(); err != nil {
		appLogger.Fatal(err)
	}
}
