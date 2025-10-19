package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/pg"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/rediseventbus"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/sse"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/subscription"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore           repository.DataStore
	KafkaAdmin          *kafka.Admin
	SSEHub              *sse.Hub
	EventBus            rediseventbus.EventBus
	InstanceID          string
	SubscriptionManager *subscription.Manager
}

// SetupGlobal initializes and returns the providers.
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

	appLogger.Info("Redis Cluster connection established",
		"addrs", cfg.Redis.Addrs,
		"cluster_mode", true)

	// Verify Redis connection with ping
	if err = redisClusterClient.Ping(ctx).Err(); err != nil {
		appLogger.Error("Redis Cluster ping failed", "error", err)

		return nil, fmt.Errorf("redis cluster ping failed: %w", err)
	}

	appLogger.Info("Redis Cluster health check passed")

	dataStore := repository.NewDataStore(pgPool)

	// Setup kafka admin
	kafkaAdmin, err := kafka.NewAdmin(&kafka.AdminConfig{
		Brokers: cfg.Kafka.Brokers,
	}, appLogger)
	if err != nil {
		appLogger.Errorf("failed to create kafka admin: %v", err)

		return nil, err
	}

	// Initialize SSE Hub
	sseHub := sse.NewHub(appLogger)

	// Start SSE Hub
	go sseHub.Run(ctx)

	// Initialize Redis for event bus
	pubSubConfig := redis.DefaultPubSubConfig()
	redisPublisher := redis.NewPublisher(redisClusterClient, pubSubConfig)
	redisSubscriber := redis.NewSubscriber(redisClusterClient, pubSubConfig, appLogger)

	appLogger.Info("Redis Pub/Sub components initialized",
		"buffer_size", pubSubConfig.ChannelBufferSize)

	// Generate instance ID
	instanceID := uuid.New().String()
	eventBus := rediseventbus.NewRedisEventBus(
		redisPublisher,
		redisSubscriber,
		instanceID,
		appLogger,
	)

	appLogger.Info("EventBus initialized with Redis sharded pub/sub",
		"instance_id", instanceID,
		"using_cluster", true)

	// Initialize SubscriptionManager for GraphQL subscriptions
	subscriptionManager := subscription.NewManager(eventBus, sseHub, appLogger)

	appLogger.Info("Subscription manager initialized for GraphQL and SSE cross-instance messaging",
		"instance_id", instanceID)

	return &Providers{
		KafkaAdmin:          kafkaAdmin,
		DataStore:           dataStore,
		SSEHub:              sseHub,
		EventBus:            eventBus,
		InstanceID:          instanceID,
		SubscriptionManager: subscriptionManager,
	}, nil
}
