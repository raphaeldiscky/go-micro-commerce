package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/notification-service/internal/server"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
)

func runKafkaWorker(ctx context.Context, cfg *config.Config, appLogger logger.Logger) {
	srv := server.NewKafkaConsumerServer(cfg)
	go func() {
		if err := srv.Start(); err != nil {
			appLogger.Errorf("Kafka server failed to start: %v", err)
		}
	}()

	<-ctx.Done()
	srv.Shutdown()
}
