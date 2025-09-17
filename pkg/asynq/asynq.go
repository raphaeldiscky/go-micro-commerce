// Package asynq is a wrapper for the asynq library.
package asynq

import (
	"time"

	"github.com/hibiken/asynq"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/config"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// Client wraps asynq.Client and provides additional functionality.
type Client struct {
	client *asynq.Client
	logger logger.Logger
}

// Server wraps asynq.Server and provides additional functionality.
type Server struct {
	server *asynq.Server
	logger logger.Logger
}

// NewClient creates a new Asynq client.
func NewClient(cfg *config.AsynqConfig, appLogger logger.Logger) (*Client, error) {
	if err := validateAsynqConfig(cfg); err != nil {
		return nil, err
	}

	// Use Redis cluster client
	redisOpt := asynq.RedisClusterClientOpt{
		Addrs:    cfg.RedisAddrs,
		Password: cfg.RedisPassword,
	}

	client := asynq.NewClient(redisOpt)

	appLogger.Printf("asynq client connected to redis cluster at %v", cfg.RedisAddrs)

	return &Client{
		client: client,
		logger: appLogger,
	}, nil
}

// NewServer creates a new Asynq server.
func NewServer(cfg *config.AsynqConfig, appLogger logger.Logger) (*Server, error) {
	if err := validateAsynqConfig(cfg); err != nil {
		return nil, err
	}

	// Use Redis cluster client
	redisOpt := asynq.RedisClusterClientOpt{
		Addrs:    cfg.RedisAddrs,
		Password: cfg.RedisPassword,
	}

	serverConfig := asynq.Config{
		Concurrency:              cfg.Concurrency,
		Queues:                   cfg.Queues,
		RetryDelayFunc:           createRetryDelayFunc(cfg),
		HealthCheckInterval:      cfg.HealthCheckInterval,
		DelayedTaskCheckInterval: cfg.DelayedTaskCheckInterval,
		Logger:                   &asynqLogger{logger: appLogger},
	}

	server := asynq.NewServer(redisOpt, serverConfig)

	appLogger.Printf("asynq server created with concurrency %d", cfg.Concurrency)

	return &Server{
		server: server,
		logger: appLogger,
	}, nil
}

// Enqueue enqueues a task to be processed immediately.
func (c *Client) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if c.client == nil {
		return nil, ErrClientNotInitialized
	}

	info, err := c.client.Enqueue(task, opts...)
	if err != nil {
		c.logger.Errorf("failed to enqueue task %s: %v", task.Type(), err)
		return nil, err
	}

	c.logger.Printf("enqueued task %s with ID %s", task.Type(), info.ID)

	return info, nil
}

// EnqueueIn enqueues a task to be processed after the given delay.
func (c *Client) EnqueueIn(
	d time.Duration,
	task *asynq.Task,
	opts ...asynq.Option,
) (*asynq.TaskInfo, error) {
	if c.client == nil {
		return nil, ErrClientNotInitialized
	}

	allOpts := make([]asynq.Option, len(opts)+1)
	copy(allOpts, opts)
	allOpts[len(opts)] = asynq.ProcessIn(d)

	info, err := c.client.Enqueue(task, allOpts...)
	if err != nil {
		c.logger.Errorf("failed to enqueue delayed task %s: %v", task.Type(), err)
		return nil, err
	}

	c.logger.Printf("enqueued delayed task %s with ID %s (delay: %v)", task.Type(), info.ID, d)

	return info, nil
}

// EnqueueAt enqueues a task to be processed at the given time.
func (c *Client) EnqueueAt(
	t time.Time,
	task *asynq.Task,
	opts ...asynq.Option,
) (*asynq.TaskInfo, error) {
	if c.client == nil {
		return nil, ErrClientNotInitialized
	}

	allOpts := make([]asynq.Option, 0, len(opts)+1)
	copy(allOpts, opts)
	allOpts[len(opts)] = asynq.ProcessAt(t)

	info, err := c.client.Enqueue(task, allOpts...)
	if err != nil {
		c.logger.Errorf("failed to enqueue scheduled task %s: %v", task.Type(), err)
		return nil, err
	}

	c.logger.Printf(
		"enqueued scheduled task %s with ID %s (scheduled for: %v)",
		task.Type(),
		info.ID,
		t,
	)

	return info, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	if c.client == nil {
		return ErrClientNotInitialized
	}

	err := c.client.Close()
	if err != nil {
		c.logger.Errorf("failed to close asynq client: %v", err)
		return err
	}

	c.logger.Printf("asynq client closed")

	return nil
}

// Start starts the task server.
func (s *Server) Start(handler asynq.Handler) error {
	if s.server == nil {
		return ErrServerNotInitialized
	}

	s.logger.Printf("starting asynq server...")

	if err := s.server.Start(handler); err != nil {
		s.logger.Errorf("failed to start asynq server: %v", err)
		return err
	}

	s.logger.Printf("asynq server started successfully")

	return nil
}

// Stop stops the task server.
func (s *Server) Stop() {
	if s.server != nil {
		s.server.Stop()
		s.logger.Printf("asynq server stopped")
	}
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() {
	if s.server != nil {
		s.server.Shutdown()
		s.logger.Printf("asynq server shutdown gracefully")
	}
}

// validateAsynqConfig validates the Asynq configuration.
func validateAsynqConfig(cfg *config.AsynqConfig) error {
	if len(cfg.RedisAddrs) == 0 {
		return ErrInvalidRedisAddr
	}

	if cfg.Concurrency <= 0 {
		return ErrInvalidConcurrency
	}

	if len(cfg.Queues) == 0 {
		return ErrInvalidQueues
	}

	if cfg.MaxRetry < 0 {
		return ErrInvalidMaxRetry
	}

	return nil
}

// createRetryDelayFunc creates a retry delay function based on config.
func createRetryDelayFunc(cfg *config.AsynqConfig) asynq.RetryDelayFunc {
	return func(n int, _ error, _ *asynq.Task) time.Duration {
		if n == 0 {
			return cfg.RetryDelay
		}

		// Exponential backoff with max delay
		delay := time.Duration(n) * cfg.RetryDelay
		if delay > cfg.RetryMaxDelay {
			return cfg.RetryMaxDelay
		}

		return delay
	}
}

// asynqLogger adapts our logger interface to asynq's logger interface.
type asynqLogger struct {
	logger logger.Logger
}

// Debug logs a debug message.
func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Printf("ASYNQ DEBUG: %v", args)
}

// Info logs an info message.
func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Printf("ASYNQ INFO: %v", args)
}

// Warn logs a warning message.
func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Printf("ASYNQ WARN: %v", args)
}

// Error logs an error message.
func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Printf("ASYNQ ERROR: %v", args)
}

// Fatal logs a fatal message.
func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Printf("ASYNQ FATAL: %v", args)
}
