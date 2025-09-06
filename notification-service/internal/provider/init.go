package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/db"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/repository"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore  repository.DataStore
	KafkaAdmin *kafka.Admin
}

// SetupGlobal initializes and returns the providers.
func SetupGlobal(cfg *config.Config) (*Providers, error) {
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

	dataStore := repository.NewDataStore(pgPool)
	// Setup kafka admin
	kafkaAdmin := kafka.NewAdmin(&kafka.AdminConfig{
		Brokers: cfg.Kafka.Brokers,
	})

	return &Providers{
		KafkaAdmin: kafkaAdmin,
		DataStore:  dataStore,
	}, nil
}
