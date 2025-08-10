package provider

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raphaeldiscky/go-micro-template/pkg/db"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/repository"
)

var (
	dbPool     *pgxpool.Pool
	dataStore  repository.DataStore
	kafkaAdmin *mq.KafkaAdmin
)

func SetupGlobal(cfg *config.Config) {
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
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pgPool.Close()
	// Setup datastore
	dataStore = repository.NewDataStore(pgPool)
	// Setup kafka admin
	kafkaAdmin = mq.NewKafkaAdmin(&mq.KafkaAdminConfig{
		Brokers: cfg.Kafka.Brokers,
	})
}
