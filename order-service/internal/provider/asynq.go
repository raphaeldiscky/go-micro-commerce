package provider

import (
	"github.com/hibiken/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgAsynq "github.com/raphaeldiscky/go-micro-commerce/pkg/asynq"
	pkgConfig "github.com/raphaeldiscky/go-micro-commerce/pkg/config"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/task"
)

// AsynqProvider holds asynq client, server and related services.
type AsynqProvider struct {
	Client                     pkgAsynq.ClientInterface
	Server                     *pkgAsynq.Server
	PaymentReminderTaskService service.PaymentReminderTaskService
	TaskHandler                *handler.PaymentReminderTaskHandler
	Mux                        *asynq.ServeMux
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

	// Create asynq client
	client, err := pkgAsynq.NewClient(asynqConfig, logger)
	if err != nil {
		return nil, err
	}

	providers.AsyncqClient = client

	// Create asynq server
	server, err := pkgAsynq.NewServer(asynqConfig, logger)
	if err != nil {
		return nil, err
	}

	// Create payment reminder task service
	paymentReminderTaskService := service.NewPaymentReminderTaskService(
		providers.NotificationRequestProducer,
		providers.DataStore,
		logger,
	)

	// Create task handler
	taskHandler := handler.NewPaymentReminderTaskHandler(
		paymentReminderTaskService,
		logger,
	)

	// Setup task routing
	mux := asynq.NewServeMux()
	mux.HandleFunc(
		task.PaymentReminderTaskType,
		taskHandler.HandlePaymentReminderTask,
	)
	mux.HandleFunc(
		task.CancelOrderTaskType,
		taskHandler.HandleCancelOrderTask,
	)

	return &AsynqProvider{
		Client:                     client,
		Server:                     server,
		PaymentReminderTaskService: paymentReminderTaskService,
		TaskHandler:                taskHandler,
		Mux:                        mux,
	}, nil
}
