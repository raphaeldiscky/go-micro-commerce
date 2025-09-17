package asynq

import (
	"github.com/hibiken/asynq"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/config"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// Inspector defines the interface for Asynq inspector operations.
type Inspector interface {
	// Close closes the inspector connection.
	Close() error
	// DeleteTask deletes a task from the queue.
	DeleteTask(queue, taskID string) error
	// ListScheduledTasks returns all scheduled tasks in the queue.
	ListScheduledTasks(queue string) ([]*asynq.TaskInfo, error)
}

// inspector wraps asynq inspector functionality.
type inspector struct {
	inspector *asynq.Inspector
	logger    logger.Logger
}

// NewInspector creates a new asynq inspector.
func NewInspector(cfg *config.AsynqConfig, logger logger.Logger) (Inspector, error) {
	redisOpt := &asynq.RedisClusterClientOpt{
		Addrs:    cfg.RedisAddrs,
		Password: cfg.RedisPassword,
	}

	ins := asynq.NewInspector(redisOpt)

	logger.Infof("asynq inspector created")

	return &inspector{
		inspector: ins,
		logger:    logger,
	}, nil
}

// Close closes the inspector connection.
func (i *inspector) Close() error {
	return i.inspector.Close()
}

// DeleteTask deletes a task from the queue.
func (i *inspector) DeleteTask(queue, taskID string) error {
	return i.inspector.DeleteTask(queue, taskID)
}

// ListScheduledTasks returns all scheduled tasks in the queue.
func (i *inspector) ListScheduledTasks(queue string) ([]*asynq.TaskInfo, error) {
	return i.inspector.ListScheduledTasks(queue)
}
