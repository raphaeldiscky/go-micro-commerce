package provider

import (
	"github.com/raphaeldiscky/go-micro-template/pkg/db"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/raphaeldiscky/go-micro-template/pkg/utils/jwtutils"

	pkgConfig "github.com/raphaeldiscky/go-micro-template/pkg/config"

	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/auth-service/internal/repository"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore  repository.DataStore
	KafkaAdmin *mq.KafkaAdmin
	JWTUtils   jwtutils.Interface
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

	jwtUtils := jwtutils.NewJWTUtils(&pkgConfig.JWTConfig{
		Secret:         cfg.JWT.Secret,
		ExpirationTime: cfg.JWT.ExpirationTime,
		RefreshTime:    cfg.JWT.RefreshTime,
		Issuer:         cfg.JWT.Issuer,
		TokenLookup:    cfg.JWT.TokenLookup,
		AuthScheme:     cfg.JWT.AuthScheme,
		SigningMethod:  cfg.JWT.SigningMethod,
		ContextKey:     cfg.JWT.ContextKey,
		AllowedAlgs:    cfg.JWT.AllowedAlgs,
	})

	// Setup datastore
	dataStore := repository.NewDataStore(pgPool)

	// Setup kafka admin
	kafkaAdmin := mq.NewKafkaAdmin(&mq.KafkaAdminConfig{
		Brokers: cfg.Kafka.Brokers,
	})

	return &Providers{
		DataStore:  dataStore,
		KafkaAdmin: kafkaAdmin,
		JWTUtils:   jwtUtils,
	}, nil
}
