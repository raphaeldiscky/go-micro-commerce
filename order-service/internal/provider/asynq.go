package provider

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgAsynq "github.com/raphaeldiscky/go-micro-commerce/pkg/asynq"
	pkgConfig "github.com/raphaeldiscky/go-micro-commerce/pkg/config"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/task"
)

// SetupAsynqClient initializes the asynq client and inspector early.
func SetupAsynqClient(
	cfg *config.Config,
	providers *Providers,
	logger logger.Logger,
) error {
	// Create asynq config with all required fields
	asynqConfig := &pkgConfig.AsynqConfig{
		RedisAddrs:               cfg.Asynq.RedisAddrs,
		RedisPassword:            cfg.Asynq.RedisPassword,
		Concurrency:              cfg.Asynq.Concurrency,
		Queues:                   cfg.Asynq.Queues,
		MaxRetry:                 cfg.Asynq.MaxRetry,
		RetryDelay:               cfg.Asynq.RetryDelay,
		RetryMaxDelay:            cfg.Asynq.RetryMaxDelay,
		HealthCheckInterval:      cfg.Asynq.HealthCheckInterval,
		DelayedTaskCheckInterval: cfg.Asynq.DelayedTaskCheckInterval,
	}

	// Create client
	client, err := pkgAsynq.NewClient(asynqConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to create asynq client: %w", err)
	}

	providers.AsynqClient = client

	// Create inspector
	inspector, err := pkgAsynq.NewInspector(asynqConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to create asynq inspector: %w", err)
	}

	providers.AsynqInspector = inspector

	return nil
}

// AsynqProvider holds asynq client, server and related services.
type AsynqProvider struct {
	Client                 pkgAsynq.Client
	Server                 pkgAsynq.Server
	PaymentReminderService service.PaymentReminderService
	TaskHandler            *handler.PaymentReminderTaskHandler
	Mux                    *asynq.ServeMux
}

// SetupAsynq initializes asynq client, server and task handlers.
func SetupAsynq(
	cfg *config.Config,
	providers *Providers,
	logger logger.Logger,
) (*AsynqProvider, error) {
	// Create asynq config
	asynqConfig := &pkgConfig.AsynqConfig{
		RedisAddrs:               cfg.Asynq.RedisAddrs,
		RedisPassword:            cfg.Asynq.RedisPassword,
		Concurrency:              cfg.Asynq.Concurrency,
		Queues:                   cfg.Asynq.Queues,
		MaxRetry:                 cfg.Asynq.MaxRetry,
		RetryDelay:               cfg.Asynq.RetryDelay,
		RetryMaxDelay:            cfg.Asynq.RetryMaxDelay,
		HealthCheckInterval:      cfg.Asynq.HealthCheckInterval,
		DelayedTaskCheckInterval: cfg.Asynq.DelayedTaskCheckInterval,
	}

	var client pkgAsynq.Client

	if providers.AsynqClient == nil {
		var err error

		client, err = pkgAsynq.NewClient(asynqConfig, logger)
		if err != nil {
			return nil, err
		}

		providers.AsynqClient = client
	} else {
		client = providers.AsynqClient
	}

	// Create asynq server
	server, err := pkgAsynq.NewServer(asynqConfig, logger)
	if err != nil {
		return nil, err
	}

	// Create payment reminder task service
	paymentReminderService := service.NewPaymentReminderService(
		providers.NotificationRequestProducer,
		providers.OrderLifecycleProducer,
		providers.DataStore,
		providers.SagaOrchestrator,
		logger,
	)

	// Create task handler
	taskHandler := handler.NewPaymentReminderTaskHandler(
		paymentReminderService,
		logger,
	)

	// Setup task routing
	mux := asynq.NewServeMux()
	mux.HandleFunc(
		task.PaymentReminderTaskType,
		taskHandler.HandlePaymentReminderTask,
	)
	mux.HandleFunc(
		task.ExpireOrderPaymentTaskType,
		taskHandler.HandleExpireOrderPaymentTask,
	)

	return &AsynqProvider{
		Client:                 client,
		Server:                 server,
		PaymentReminderService: paymentReminderService,
		TaskHandler:            taskHandler,
		Mux:                    mux,
	}, nil
}
