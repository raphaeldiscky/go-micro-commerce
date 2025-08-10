package worker

import (
	"context"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/server"
)

func runHTTPWorker(cfg *config.Config, lgr logger.Logger, ctx context.Context) {
	srv := server.NewHTTPServer(cfg, lgr)
	go srv.Start()

	<-ctx.Done()
	srv.Shutdown()
}
