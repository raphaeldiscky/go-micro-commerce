package provider

import (
	"context"

	"github.com/bsm/redislock"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/asynq"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/pg"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/service"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore      repository.DataStore
	KafkaAdmin     *kafka.Admin
	ProductClient  client.ProductClient
	AsynqClient    asynq.Client
	AsynqInspector asynq.Inspector

	NotificationRequestProducer    kafka.Producer
	CartService                    service.CartService
	CheckoutSessionService         service.CheckoutSessionService
	CheckoutSessionReminderService service.CheckoutSessionReminderService
}

// SetupGlobal initializes all providers.
func SetupGlobal(
	ctx context.Context,
	cfg *config.Config,
	appLogger logger.Logger,
	tel *telemetry.Telemetry,
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

	// Setup kafka admin
	kafkaAdmin, err := kafka.NewAdmin(&kafka.AdminConfig{
		Brokers: cfg.Kafka.Brokers,
	}, appLogger)
	if err != nil {
		appLogger.Errorf("failed to create kafka admin: %v", err)

		return nil, err
	}

	// Setup redis lock client
	redisLockClient := redislock.New(redisClusterClient)

	// Setup datastore
	dataStore := repository.NewDataStore(pgPool, redisLockClient, appLogger, tel)

	// Setup product client
	productClient, err := client.NewProductClient(cfg)
	if err != nil {
		appLogger.Errorf("failed to create product client: %v", err)

		return nil, err
	}

	return &Providers{
		DataStore:     dataStore,
		KafkaAdmin:    kafkaAdmin,
		ProductClient: productClient,
	}, nil
}
