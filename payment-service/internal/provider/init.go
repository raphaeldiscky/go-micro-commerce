package provider

import (
	"context"

	"github.com/bsm/redislock"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/pg"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/gateway"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/job"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/service"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore             repository.DataStore
	KafkaAdmin            *kafka.Admin
	PaymentGatewayClients map[string]client.PaymentGatewayClient
	PaymentService        service.PaymentService
	JobScheduler          *job.Scheduler
}

// SetupGlobal initializes all providers.
func SetupGlobal(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
) (*Providers, error) {
	pgPool, err := pg.NewPostgresConnection(ctx, &pg.PostgresConfig{
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

	// Setup payment gateway clients using factory
	gatewayFactory := gateway.NewFactory(cfg.PaymentGateway, appLogger)
	paymentGatewayClients := gatewayFactory.CreateGateways()

	providers := &Providers{
		DataStore:             dataStore,
		KafkaAdmin:            kafkaAdmin,
		PaymentGatewayClients: paymentGatewayClients,
	}

	// Setup job scheduler with payment timeout job
	SetupJobScheduler(cfg, appLogger, providers)

	return providers, nil
}

// SetupJobScheduler initializes the job scheduler with registered jobs.
func SetupJobScheduler(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) {
	// Create job scheduler
	scheduler := job.NewScheduler(appLogger, cfg.Job)

	// Initialize payment service (required for timeout job)
	paymentService := service.NewPaymentService(
		providers.DataStore,
		appLogger,
		providers.PaymentGatewayClients,
	)
	providers.PaymentService = paymentService

	// Create and register payment timeout job
	if cfg.Job.PaymentTimeout.Enabled {
		timeoutJob := job.NewPaymentTimeoutJob(
			paymentService,
			providers.DataStore,
			cfg,
			appLogger,
			cfg.Job.PaymentTimeout.Interval,
		)
		scheduler.RegisterJob(timeoutJob)
		appLogger.Info("Payment timeout job registered")
	} else {
		appLogger.Info("Payment timeout job is disabled")
	}

	providers.JobScheduler = scheduler
}
