// Package integration_test provides integration tests for the chat service.
package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bsm/redislock"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/eventbus"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
	"github.com/stretchr/testify/suite"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/server"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/service"
	chatwebsocket "github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

const (
	debugLevel         = 4
	httpRequestTimeout = 15 * time.Second
	wsHandshakeTimeout = 10 * time.Second
)

// TestSuite holds all integration tests.
type TestSuite struct {
	suite.Suite

	tcSetup    *TestContainersSetup
	ctx        context.Context
	cancelFunc context.CancelFunc
	logger     logger.Logger
}

// ChatServiceInstance represents a running chat-service instance.
type ChatServiceInstance struct {
	Port         int
	InstanceID   string
	Hub          *chatwebsocket.ChatHub
	EventBus     eventbus.EventBus
	Server       *server.WebSocketServer
	DataStore    repository.DataStore
	BaseURL      string
	WebSocketURL string
	shutdownFunc context.CancelFunc
}

// SetupSuite runs once before all tests.
func (s *TestSuite) SetupSuite() {
	s.ctx, s.cancelFunc = context.WithCancel(context.Background())

	// Setup logger
	s.logger = logger.NewLogrusLogger(debugLevel)

	// Setup testcontainers
	s.tcSetup = NewTestContainersSetup()

	err := s.tcSetup.SetupPostgres()
	s.Require().NoError(err, "Failed to setup PostgreSQL")

	err = s.tcSetup.SetupRedis()
	s.Require().NoError(err, "Failed to setup Redis")

	s.logger.Info("Test infrastructure ready",
		"postgres", "running",
		"redis", "running")
}

// TearDownSuite runs once after all tests.
func (s *TestSuite) TearDownSuite() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	if s.tcSetup != nil {
		s.tcSetup.Cleanup()
	}
}

// SetupTest runs before each test.
func (s *TestSuite) SetupTest() {
	// Clean up data before each test
	if s.tcSetup != nil && s.tcSetup.DBPool != nil {
		err := s.tcSetup.CleanupData()
		s.Require().NoError(err, "Failed to cleanup data")
	}
}

// StartChatServiceInstance starts a new chat-service instance on the specified port.
func (s *TestSuite) StartChatServiceInstance(port int) (*ChatServiceInstance, error) {
	// Get connection strings
	pgConnStr, err := s.tcSetup.GetPostgresConnectionString()
	if err != nil {
		return nil, err
	}

	// Create database pool using connection string
	pgPool, err := pgxpool.New(s.ctx, pgConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Create Redis lock client
	lockClient := redislock.New(s.tcSetup.RedisClient)

	// Create data store
	dataStore := repository.NewDataStore(pgPool, lockClient, s.logger)

	// Create Redis pub/sub clients
	pubSubConfig := redis.DefaultPubSubConfig()
	redisPublisher := redis.NewPublisher(s.tcSetup.RedisClient, pubSubConfig)
	redisSubscriber := redis.NewSubscriber(s.tcSetup.RedisClient, pubSubConfig, s.logger)

	// Create EventBus
	instanceID := uuid.New().String()
	eventBus := eventbus.NewRedisEventBus(
		redisPublisher,
		redisSubscriber,
		instanceID,
		s.logger,
	)

	// Create WebSocket hub
	hub := chatwebsocket.NewChatHub(
		dataStore.ConnectionRepository(),
		dataStore.MessageRepository(),
		s.logger,
		instanceID,
	)

	// Set EventBus on hub
	hub.SetEventBus(eventBus)

	// Create test config
	cfg := &config.Config{
		WebSocketServer: &config.WebSocketServerConfig{
			Port:         port,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  120 * time.Second,
			HSTSMaxAge:   31536000,
			RateLimiter:  100,
		},
	}

	// Create services
	chatService := service.NewChatService(
		dataStore,
		s.logger,
		hub,
	)

	// Create minimal connection service config for testing
	nodeConfig := &service.NodeConfig{
		DefaultNodeAddress: "",   // not needed for testing
		MaxConnections:     1000, // default max connections
		ConsulAddress:      "",   // not needed for testing
		ChatServiceName:    "chat-service-test",
	}

	connectionService := service.NewConnectionService(
		s.logger,
		"../../keys/public.pem", // publicKeyPath for JWT validation
		"",                      // jwksURL - not needed for testing
		0,                       // jwksCacheTTL
		0,                       // jwksRefreshInterval
		nodeConfig,
	)

	// Create WebSocket server
	instanceCtx, cancelFunc := context.WithCancel(s.ctx)

	wsServer := server.NewWebSocketServer(
		hub,
		cfg,
		s.logger,
		connectionService,
		chatService,
	)

	// Start server in goroutine
	go func() {
		if err = wsServer.Start(instanceCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("WebSocket server error",
				"port", port,
				"error", err)
		}
	}()

	// Wait for server to be ready
	time.Sleep(500 * time.Millisecond)

	instance := &ChatServiceInstance{
		Port:         port,
		InstanceID:   instanceID,
		Hub:          hub,
		EventBus:     eventBus,
		Server:       wsServer,
		DataStore:    dataStore,
		BaseURL:      fmt.Sprintf("http://localhost:%d", port),
		WebSocketURL: fmt.Sprintf("ws://localhost:%d/ws", port),
		shutdownFunc: cancelFunc,
	}

	s.logger.Info("Chat service instance started",
		"port", port,
		"instance_id", instanceID)

	return instance, nil
}

// StopChatServiceInstance stops a chat-service instance.
func (s *TestSuite) StopChatServiceInstance(instance *ChatServiceInstance) error {
	if instance.shutdownFunc != nil {
		instance.shutdownFunc()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := instance.Hub.Shutdown(shutdownCtx); err != nil {
		return err
	}

	s.logger.Info("Chat service instance stopped",
		"port", instance.Port,
		"instance_id", instance.InstanceID)

	return nil
}

// ConnectWebSocket creates a WebSocket connection to a chat-service instance.
func (s *TestSuite) ConnectWebSocket(
	instance *ChatServiceInstance,
	userID uuid.UUID,
	userType constant.UserType,
) (*websocket.Conn, error) {
	// Create WebSocket dialer
	dialer := websocket.Dialer{
		HandshakeTimeout: wsHandshakeTimeout,
	}

	// Create request headers with auth info
	headers := http.Header{}
	headers.Set("X-User-Id", userID.String())
	headers.Set("X-User-Type", string(userType))

	// Connect to WebSocket
	conn, resp, err := dialer.Dial(instance.WebSocketURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to connect WebSocket: %w", err)
	}

	if resp != nil {
		_ = resp.Body.Close()
	}

	return conn, nil
}

// CreateTestConversation creates a test conversation in the database.
func (s *TestSuite) CreateTestConversation(
	instance *ChatServiceInstance,
	participants []uuid.UUID,
) (*entity.Conversation, error) {
	conversationID := uuid.New()

	// Create conversation
	conversation := &entity.Conversation{
		ID:        conversationID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := instance.DataStore.ConversationRepository().Create(s.ctx, conversation)
	if err != nil {
		return nil, err
	}

	// Add participants
	for _, participantID := range participants {
		participant := &entity.Participant{
			ConversationID: conversationID,
			UserID:         participantID,
			UserType:       constant.UserTypeUser,
			JoinedAt:       time.Now(),
		}

		if _, err = instance.DataStore.ParticipantRepository().Create(s.ctx, participant); err != nil {
			return nil, err
		}
	}

	return conversation, nil
}

// SendWebSocketMessage sends a message through WebSocket connection.
func (s *TestSuite) SendWebSocketMessage(
	conn *websocket.Conn,
	msgType pkgwebsocket.MessageType,
	payload interface{},
) error {
	msg := pkgwebsocket.Message{
		ID:        uuid.New(),
		Type:      msgType,
		Timestamp: time.Now(),
	}

	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		msg.Content = json.RawMessage(payloadBytes)
	}

	return conn.WriteJSON(msg)
}

// ReceiveWebSocketMessage receives a message from WebSocket connection with timeout.
func (s *TestSuite) ReceiveWebSocketMessage(
	conn *websocket.Conn,
	timeout time.Duration,
) (*pkgwebsocket.Message, error) {
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	var msg pkgwebsocket.Message
	if err := conn.ReadJSON(&msg); err != nil {
		return nil, err
	}

	return &msg, nil
}
