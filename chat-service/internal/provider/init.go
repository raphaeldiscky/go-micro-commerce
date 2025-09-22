package provider

import (
	"context"

	"github.com/bsm/redislock"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/db"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"

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
	WebSocketHub         *websocket.ChatHub
	RedisPublisher       redis.Publisher
	RedisSubscriber      redis.Subscriber
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

	// Create Redis pub/sub clients for chat service
	pubSubConfig := redis.DefaultPubSubConfig()
	redisPublisher := redis.NewPublisher(redisClusterClient, pubSubConfig)
	redisSubscriber := redis.NewSubscriber(redisClusterClient, pubSubConfig, appLogger)

	return &Providers{
		DataStore:            dataStore,
		ConnectionRepository: dataStore.ConnectionRepository(),
		ChatService:          nil, // Will be set in SetupChat
		WebSocketHub:         nil, // Will be set in SetupChat
		RedisPublisher:       redisPublisher,
		RedisSubscriber:      redisSubscriber,
	}, nil
}
