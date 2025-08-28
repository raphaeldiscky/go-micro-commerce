package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/server"
)

func runGRPCWorker(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) {
	srv := server.NewGRPCServer(providers.ProductService, appLogger, cfg)
	go func() {
		if err := srv.StartGRPC(); err != nil {
			appLogger.Errorf("gRPC server failed to start: %v", err)
		}
	}()
	<-ctx.Done()
	srv.Shutdown()
}
