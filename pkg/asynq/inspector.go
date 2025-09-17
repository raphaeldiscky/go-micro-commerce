package asynq

import (
	"github.com/hibiken/asynq"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/config"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// Inspector wraps asynq inspector functionality.
type Inspector struct {
	*asynq.Inspector
}

// NewInspector creates a new asynq inspector.
func NewInspector(cfg *config.AsynqConfig, logger logger.Logger) (*Inspector, error) {
	redisOpt := &asynq.RedisClusterClientOpt{
		Addrs:    cfg.RedisAddrs,
		Password: cfg.RedisPassword,
	}

	inspector := asynq.NewInspector(redisOpt)

	logger.Infof("asynq inspector created")

	return &Inspector{
		Inspector: inspector,
	}, nil
}

// Close closes the inspector connection.
func (i *Inspector) Close() error {
	return i.Inspector.Close()
}
