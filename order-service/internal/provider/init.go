package provider

import (
	"context"

	"github.com/bsm/redislock"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/db"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/temporal"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/job"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/service"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore                   repository.DataStore
	KafkaAdmin                  *kafka.Admin
	JobScheduler                *job.Scheduler
	TemporalClient              *client.TemporalClient
	FulfillmentClient           client.FulfillmentClientInterface
	PaymentClient               client.PaymentClientInterface
	NotificationRequestProducer kafka.ProducerInterface
	ReminderScheduler           *temporal.ReminderScheduler
	PaymentReminderService      *service.PaymentReminderService
	AsyncqClient                *asynq.Client
}

// SetupGlobal initializes all providers.
func SetupGlobal(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
) (*Providers, error) {
	pgPool, err := db.NewPostgresConnection(ctx, &db.PostgresConfig{
		Host:            cfg.Postgres.Host,
		Port:            cfg.Postgres.Port,
		User:            cfg.Postgres.User,
		Password:        cfg.Postgres.Password,
		DB:              cfg.Postgres.DB,
		SSLMode:         cfg.Postgres.SSLMode,
		MaxIdleConns:    cfg.Postgres.MaxIdleConns,
		MaxOpenConns:    cfg.Postgres.MaxOpenConns,
		MaxConnLifetime: cfg.Postgres.MaxConnLifetime,
	}, appLogger)
	if err != nil {
		return nil, err
	}

	redisClusterClient, err := redis.NewRedisCluster(ctx, &redis.ClusterConfig{
		Addrs:           cfg.Redis.Addrs,
		Password:        cfg.Redis.Password,
		DialTimeout:     cfg.Redis.DialTimeout,
		ReadTimeout:     cfg.Redis.ReadTimeout,
		WriteTimeout:    cfg.Redis.WriteTimeout,
		MinIdleConn:     cfg.Redis.MinIdleConn,
		MaxIdleConn:     cfg.Redis.MaxIdleConn,
		MaxActiveConn:   cfg.Redis.MaxActiveConn,
		MaxConnLifetime: cfg.Redis.MaxConnLifetime,
	}, appLogger)
	if err != nil {
		return nil, err
	}

	lockClient := redislock.New(redisClusterClient)
	dataStore := repository.NewDataStore(pgPool, lockClient, appLogger)
	// Setup kafka admin
	kafkaAdmin, err := kafka.NewAdmin(&kafka.AdminConfig{
		Brokers: cfg.Kafka.Brokers,
	}, appLogger)
	if err != nil {
		appLogger.Errorf("failed to create kafka admin: %v", err)

		return nil, err
	}

	// Setup fulfillment client for event correlation
	fulfillmentClient, err := client.NewFulfillmentClient(
		cfg,
		appLogger,
	) // logger will be injected later
	if err != nil {
		appLogger.Warnf(
			"failed to create fulfillment client: %v. Order service will start without fulfillment client functionality.",
			err,
		)

		fulfillmentClient = nil
	}

	// Setup payment client for event correlation
	paymentClient := client.NewPaymentClient(appLogger)

	return &Providers{
		DataStore:                   dataStore,
		KafkaAdmin:                  kafkaAdmin,
		TemporalClient:              nil, // will be set up later in worker
		FulfillmentClient:           fulfillmentClient,
		PaymentClient:               paymentClient,
		NotificationRequestProducer: nil,
	}, nil
}
