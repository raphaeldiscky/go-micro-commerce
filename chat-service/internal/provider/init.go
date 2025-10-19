package provider

import (
	"context"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/pg"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/rediseventbus"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/service"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore            repository.DataStore
	ConnectionRepository repository.ConnectionRepository
	ChatService          service.ChatService
	ConnectionService    service.ConnectionService
	WebSocketHub         *websocket.ChatHub
	EventBus             rediseventbus.EventBus
	Logger               logger.Logger
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

	// Create Redis pub/sub clients for event bus
	pubSubConfig := redis.DefaultPubSubConfig()
	redisPublisher := redis.NewPublisher(redisClusterClient, pubSubConfig)
	redisSubscriber := redis.NewSubscriber(redisClusterClient, pubSubConfig, appLogger)

	// Create EventBus for cross-instance messaging
	instanceID := uuid.New().String()
	eventBus := rediseventbus.NewRedisEventBus(
		redisPublisher,
		redisSubscriber,
		instanceID,
		appLogger,
	)

	// Initialize WebSocket hub
	webSocketHub := initWebSocketHub(
		dataStore.ConnectionRepository(),
		dataStore.MessageRepository(),
		appLogger,
		instanceID,
	)

	// Set EventBus on hub (must be done after hub creation)
	webSocketHub.SetEventBus(eventBus)

	return &Providers{
		DataStore:            dataStore,
		ConnectionRepository: dataStore.ConnectionRepository(),
		ChatService:          nil, // Will be set in SetupChat
		ConnectionService:    nil, // Will be set in SetupChat
		WebSocketHub:         webSocketHub,
		EventBus:             eventBus,
		Logger:               appLogger,
	}, nil
}

// initWebSocketHub initializes the WebSocket hub.
func initWebSocketHub(
	connectionRepo repository.ConnectionRepository,
	messageRepo repository.MessageRepository,
	logger logger.Logger,
	instanceID string,
) *websocket.ChatHub {
	return websocket.NewChatHub(connectionRepo, messageRepo, logger, instanceID)
}
