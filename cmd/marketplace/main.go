package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"google.golang.org/grpc"

	"github.com/raphaeldiscky/go-ddd-template/internal/application/interfaces"
	"github.com/raphaeldiscky/go-ddd-template/internal/application/services"
	"github.com/raphaeldiscky/go-ddd-template/internal/domain/events"
	"github.com/raphaeldiscky/go-ddd-template/internal/infrastructure/cache"
	postgres2 "github.com/raphaeldiscky/go-ddd-template/internal/infrastructure/db/postgres"
	"github.com/raphaeldiscky/go-ddd-template/internal/infrastructure/messaging/kafka"
	"github.com/raphaeldiscky/go-ddd-template/internal/infrastructure/repository"
	"github.com/raphaeldiscky/go-ddd-template/internal/interface/api/rest"
	grpcHandlers "github.com/raphaeldiscky/go-ddd-template/internal/interface/grpc"
	marketplacev1 "github.com/raphaeldiscky/go-ddd-template/proto"
)

// Config holds the application configuration.
type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	Server   ServerConfig
}

// DatabaseConfig holds the configuration for the database connection.
type DatabaseConfig struct {
	DSN string
}

// RedisConfig holds the configuration for Redis cache.
type RedisConfig struct {
	Host       string
	Port       int
	Password   string
	DB         int
	KeyPrefix  string
	DefaultTTL time.Duration
}

// KafkaConfig holds the configuration for Kafka messaging.
type KafkaConfig struct {
	Brokers       []string
	TopicPrefix   string
	RetryAttempts int
	RetryDelay    time.Duration
}

// ServerConfig holds the configuration for the HTTP and gRPC servers.
type ServerConfig struct {
	HTTPPort string
	GRPCPort string
}

func main() {
	// Configuration - in production, load from environment variables or config files
	config := Config{
		Database: DatabaseConfig{
			DSN: "host=localhost user=marketplace password=marketplace dbname=marketplace port=9920 sslmode=disable TimeZone=Asia/Shanghai",
		},
		Redis: RedisConfig{
			Host:       "localhost",
			Port:       6379,
			Password:   "",
			DB:         0,
			KeyPrefix:  "marketplace",
			DefaultTTL: 30 * time.Minute,
		},
		Kafka: KafkaConfig{
			Brokers:       []string{"localhost:9092"},
			TopicPrefix:   "marketplace",
			RetryAttempts: 3,
			RetryDelay:    time.Second,
		},
		Server: ServerConfig{
			HTTPPort: ":8080",
			GRPCPort: ":8090",
		},
	}

	// Initialize database connection pool
	dbPool, err := postgres2.NewConnection(config.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Run database migrations
	log.Println("Running database migrations...")

	migrationConfig := postgres2.MigrationConfig{
		DatabaseURL:    config.Database.DSN,
		MigrationsPath: "./migrations",
	}
	if err := postgres2.RunMigrations(dbPool, migrationConfig); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migrations completed successfully!")

	// Initialize Redis cache
	redisCache := cache.NewRedisCache(cache.RedisConfig{
		Host:       config.Redis.Host,
		Port:       config.Redis.Port,
		Password:   config.Redis.Password,
		DB:         config.Redis.DB,
		KeyPrefix:  config.Redis.KeyPrefix,
		DefaultTTL: config.Redis.DefaultTTL,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisCache.Ping(ctx); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		log.Println("Continuing without cache...")

		redisCache = nil
	}

	// Initialize Kafka event publisher
	var kafkaPublisher events.EventPublisher

	if kafkaEventPublisher, err := kafka.NewKafkaEventPublisher(kafka.KafkaConfig{
		Brokers:       config.Kafka.Brokers,
		TopicPrefix:   config.Kafka.TopicPrefix,
		RetryAttempts: config.Kafka.RetryAttempts,
		RetryDelay:    config.Kafka.RetryDelay,
	}); err != nil {
		log.Printf("Warning: Kafka connection failed: %v", err)
		log.Println("Continuing without event publishing...")
	} else {
		kafkaPublisher = kafkaEventPublisher
	}

	// Initialize repositories
	baseProductRepo := postgres2.NewSqlcProductRepository(dbPool)
	baseSellerRepo := postgres2.NewSqlcSellerRepository(dbPool)

	// Decorate repositories with caching if Redis is available
	productRepo := baseProductRepo
	sellerRepo := baseSellerRepo

	if redisCache != nil {
		productRepo = repository.NewCachedProductRepository(
			baseProductRepo,
			redisCache,
			config.Redis.DefaultTTL,
		)
		sellerRepo = repository.NewCachedSellerRepository(
			baseSellerRepo,
			redisCache,
			config.Redis.DefaultTTL,
		)
	}

	// Initialize services
	productService := services.NewProductService(productRepo, sellerRepo, kafkaPublisher)
	sellerService := services.NewSellerService(sellerRepo, kafkaPublisher)

	// Start gRPC server
	go startGRPCServer(config.Server.GRPCPort, productService, sellerService)

	// Start Kafka consumer (optional - for demonstrating event handling)
	if kafkaPublisher != nil {
		go startKafkaConsumer(config.Kafka)
	}

	// Start HTTP server
	startHTTPServer(config.Server.HTTPPort, productService, sellerService)
}

func startGRPCServer(
	port string,
	productService interfaces.ProductService,
	sellerService interfaces.SellerService,
) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	s := grpc.NewServer()

	// Register gRPC services
	productGRPCService := grpcHandlers.NewProductServiceServer(productService)
	sellerGRPCService := grpcHandlers.NewSellerServiceServer(sellerService)

	marketplacev1.RegisterProductServiceServer(s, productGRPCService)
	marketplacev1.RegisterSellerServiceServer(s, sellerGRPCService)

	log.Printf("gRPC server starting on port %s", port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

func startHTTPServer(
	port string,
	productService interfaces.ProductService,
	sellerService interfaces.SellerService,
) {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Register REST controllers
	rest.NewProductController(e, productService)
	rest.NewSellerController(e, sellerService)

	log.Printf("HTTP server starting on port %s", port)

	if err := e.Start(port); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func startKafkaConsumer(config KafkaConfig) {
	// Example Kafka consumer setup
	topics := []string{
		config.TopicPrefix + ".ProductCreated",
		config.TopicPrefix + ".SellerCreated",
	}

	consumer, err := kafka.NewKafkaEventSubscriber(
		config.Brokers,
		"marketplace-consumer-group",
		topics,
	)
	if err != nil {
		log.Printf("Failed to create Kafka consumer: %v", err)

		return
	}

	ctx := context.Background()

	// Subscribe to events with example handlers
	err = consumer.Subscribe(
		ctx,
		"ProductCreated",
		func(_ context.Context, event events.DomainEvent) error {
			log.Printf("Received ProductCreated event: %+v", event)
			// Add your business logic here (e.g., update search index, send notifications, etc.)
			return nil
		},
	)
	if err != nil {
		log.Printf("Failed to subscribe to ProductCreated events: %v", err)
	}

	err = consumer.Subscribe(
		ctx,
		"SellerCreated",
		func(_ context.Context, event events.DomainEvent) error {
			log.Printf("Received SellerCreated event: %+v", event)
			// Add your business logic here
			return nil
		},
	)
	if err != nil {
		log.Printf("Failed to subscribe to SellerCreated events: %v", err)
	}

	// Start consuming
	if err := consumer.Start(ctx); err != nil {
		log.Printf("Failed to start Kafka consumer: %v", err)
	}
}
