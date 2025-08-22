package provider

import (
	"github.com/raphaeldiscky/go-micro-template/pkg/db"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/repository"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore  repository.DataStore
	KafkaAdmin *mq.KafkaAdmin
}

// SetupGlobal initializes all providers.
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
	kafkaAdmin := mq.NewKafkaAdmin(&mq.KafkaAdminConfig{
		Brokers: cfg.Kafka.Brokers,
	})

	return &Providers{
		DataStore:  dataStore,
		KafkaAdmin: kafkaAdmin,
	}, nil
}
