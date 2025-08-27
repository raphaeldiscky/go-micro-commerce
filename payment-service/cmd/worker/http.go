package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/server"
)

func runHTTPWorker(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) {
	srv := server.NewHTTPServer(cfg, appLogger, providers)
	go func() {
		if err := srv.Start(); err != nil {
			appLogger.Errorf("HTTP server failed to start: %v", err)
		}
	}()

	<-ctx.Done()
	srv.Shutdown()
}
