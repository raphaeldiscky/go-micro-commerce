package provider

import (
	"context"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/eventbus"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/pg"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/sharding"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/sse"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/notification"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/repository"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore     repository.DataStore
	KafkaAdmin    *kafka.Admin
	SSEHub        *sse.Hub
	EventBus      eventbus.EventBus
	InstanceID    string
	ShardResolver *sharding.ShardResolver
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

	// Generate instance ID
	instanceID := uuid.New().String()
	eventBus := eventbus.NewRedisEventBus(
		redisPublisher,
		redisSubscriber,
		instanceID,
		appLogger,
	)

	appLogger.Info("EventBus initialized with Redis pub/sub",
		"instance_id", instanceID)

	// Initialize ShardResolver with consistent hashing
	shardResolver, err := sharding.NewShardResolver(sharding.Config{
		ShardCount:        cfg.Sharding.ShardCount,
		ReplicationFactor: cfg.Sharding.ReplicationFactor,
		LoadFactor:        cfg.Sharding.LoadFactor,
	}, appLogger)
	if err != nil {
		appLogger.Errorf("failed to create shard resolver: %v", err)
		return nil, err
	}

	appLogger.Info("ShardResolver initialized with consistent hashing",
		"shard_count", cfg.Sharding.ShardCount,
		"replication_factor", cfg.Sharding.ReplicationFactor,
		"load_factor", cfg.Sharding.LoadFactor)

	// Set up SSE Hub with EventBus for cross-instance notifications
	if errSetup := setupSSEEventBus(sseHub, eventBus, instanceID, appLogger); errSetup != nil {
		appLogger.Errorf("failed to set up SSE EventBus: %v", errSetup)
		return nil, errSetup
	}

	return &Providers{
		KafkaAdmin:    kafkaAdmin,
		DataStore:     dataStore,
		SSEHub:        sseHub,
		EventBus:      eventBus,
		InstanceID:    instanceID,
		ShardResolver: shardResolver,
	}, nil
}

// setupSSEEventBus configures the SSE Hub to receive cross-instance notifications via Redis.
func setupSSEEventBus(
	sseHub *sse.Hub,
	eventBus eventbus.EventBus,
	instanceID string,
	appLogger logger.Logger,
) error {
	// Create notification event handler
	notificationEventHandler := notification.NewEventHandler(appLogger)

	// Register handler for notification created events
	notificationEventHandler.SetNotificationCreatedHandler(
		func(_ context.Context, event *notification.CreatedEvent) error {
			// Application-layer filtering: check if user is connected to this instance
			connections := sseHub.GetUserConnections(event.UserID)
			if len(connections) == 0 {
				appLogger.Debug("User not connected to this instance, skipping",
					"user_id", event.UserID)

				return nil
			}

			appLogger.Debug("Broadcasting notification to user connections",
				"user_id", event.UserID,
				"connection_count", len(connections))

			// Broadcast to all user connections on this instance
			return sseHub.BroadcastToUser(event.UserID, event.Message)
		},
	)

	// Wrap the notification handler with eventbus.EventHandler signature
	eventHandler := func(ctx context.Context, event eventbus.Event) error {
		return notificationEventHandler.HandleEvent(ctx, event)
	}

	// Configure SSE Hub with EventBus
	if err := sseHub.SetEventBus(
		eventBus,
		instanceID,
		redis.NotificationShardChannel,
		pkgconstant.SSEShardCount,
		eventHandler,
	); err != nil {
		return err
	}

	appLogger.Info("SSE Hub configured with EventBus for cross-instance notifications",
		"shard_count", pkgconstant.SSEShardCount)

	return nil
}
