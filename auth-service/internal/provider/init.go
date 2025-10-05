package provider

import (
	"context"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/db"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/kafka"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/encryptutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/jwtutils"

	pkgConfig "github.com/raphaeldiscky/go-micro-commerce/pkg/config"

	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/auth-service/internal/repository"
)

// Providers holds all initialized providers.
type Providers struct {
	DataStore    repository.DataStore
	KafkaAdmin   *kafka.Admin
	JWTUtils     jwtutils.JWT
	BcryptHasher encryptutils.Hasher
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
		DB:              cfg.Postgres.DB,
		User:            cfg.Postgres.User,
		Password:        cfg.Postgres.Password,
		SSLMode:         cfg.Postgres.SSLMode,
		MaxIdleConns:    cfg.Postgres.MaxIdleConns,
		MaxOpenConns:    cfg.Postgres.MaxOpenConns,
		MaxConnLifetime: cfg.Postgres.MaxConnLifetime,
	}, appLogger)
	if err != nil {
		return nil, err
	}

	jwtUtils := jwtutils.NewJWTUtils(&pkgConfig.JWTConfig{
		PublicKeyPath:  cfg.JWT.PublicKeyPath,
		PrivateKeyPath: cfg.JWT.PrivateKeyPath,
		ExpirationTime: cfg.JWT.ExpirationTime,
		RefreshTime:    cfg.JWT.RefreshTime,
		Issuer:         cfg.JWT.Issuer,
		TokenLookup:    cfg.JWT.TokenLookup,
		AuthScheme:     cfg.JWT.AuthScheme,
		SigningMethod:  cfg.JWT.SigningMethod,
		ContextKey:     cfg.JWT.ContextKey,
		AllowedAlgs:    cfg.JWT.AllowedAlgs,
	},
		appLogger,
	)

	bcryptHasher := encryptutils.NewBcryptHasher(cfg.Bcrypt.Cost)

	// Setup datastore
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
		DataStore:    dataStore,
		KafkaAdmin:   kafkaAdmin,
		JWTUtils:     jwtUtils,
		BcryptHasher: bcryptHasher,
	}, nil
}
