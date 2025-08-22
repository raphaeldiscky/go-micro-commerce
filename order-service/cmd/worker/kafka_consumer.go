package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"

	"github.com/raphaeldiscky/go-micro-template/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-template/order-service/internal/server"
)

func runKafkaConsumerWorker(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
	providers *provider.Providers,
) {
	srv := server.NewKafkaConsumerServer(cfg, appLogger, providers)
	go func() {
		if err := srv.Start(); err != nil {
			appLogger.Errorf("Kafka server failed to start: %v", err)
		}
	}()

	<-ctx.Done()
	srv.Shutdown()
}
