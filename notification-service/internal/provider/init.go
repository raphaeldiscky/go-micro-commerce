package provider

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/db"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/repository"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore  repository.DataStore
	KafkaAdmin *kafka.Admin
}

// SetupGlobal initializes and returns the providers.
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

	dataStore := repository.NewDataStore(pgPool)
	// Setup kafka admin
	kafkaAdmin, err := kafka.NewAdmin(&kafka.AdminConfig{
		Brokers: cfg.Kafka.Brokers,
	}, appLogger)
	if err != nil {
		appLogger.Errorf("failed to create kafka admin: %v", err)

		return nil, err
	}

	return &Providers{
		KafkaAdmin: kafkaAdmin,
		DataStore:  dataStore,
	}, nil
}
