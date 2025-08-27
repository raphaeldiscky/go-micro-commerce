package provider

import (
	"context"

	"github.com/bsm/redislock"
	"github.com/raphaeldiscky/go-micro-template/pkg/db"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/repository"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore  repository.DataStore
	KafkaAdmin *mq.KafkaAdmin
}

// SetupGlobal initializes all providers.
func SetupGlobal(ctx context.Context, cfg *config.Config) (*Providers, error) {
	pgPool, err := db.NewPostgresConnection(&db.PostgresConfig{
		Host:            cfg.Postgres.Host,
		Port:            cfg.Postgres.Port,
		User:            cfg.Postgres.User,
		Password:        cfg.Postgres.Password,
		Name:            cfg.Postgres.Name,
		SSLMode:         cfg.Postgres.SSLMode,
		MaxIdleConns:    cfg.Postgres.MaxIdleConns,
		MaxOpenConns:    cfg.Postgres.MaxOpenConns,
		MaxConnLifetime: cfg.Postgres.MaxConnLifetime,
	})
	if err != nil {
		return nil, err
	}

	redisClusterClient, err := db.NewRedisCluster(ctx, &db.RedisClusterConfig{
		Addrs:           cfg.Redis.Addrs,
		Password:        cfg.Redis.Password,
		DialTimeout:     cfg.Redis.DialTimeout,
		ReadTimeout:     cfg.Redis.ReadTimeout,
		WriteTimeout:    cfg.Redis.WriteTimeout,
		MinIdleConn:     cfg.Redis.MinIdleConn,
		MaxIdleConn:     cfg.Redis.MaxIdleConn,
		MaxActiveConn:   cfg.Redis.MaxActiveConn,
		MaxConnLifetime: cfg.Redis.MaxConnLifetime,
	})
	if err != nil {
		return nil, err
	}

	redisClient := redislock.New(redisClusterClient)
	dataStore := repository.NewDataStore(pgPool, redisClient)
	// Setup kafka admin
	kafkaAdmin := mq.NewKafkaAdmin(&mq.KafkaAdminConfig{
		Brokers: cfg.Kafka.Brokers,
	})

	return &Providers{
		DataStore:  dataStore,
		KafkaAdmin: kafkaAdmin,
	}, nil
}
