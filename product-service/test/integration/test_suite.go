// Package integration provides integration tests for the product service.
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/server"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/service"
)

// mockKafkaProducer is a mock implementation of KafkaProducerInterface for testing.
type mockKafkaProducer struct{}

func (m *mockKafkaProducer) Send(_ context.Context, _ mq.BaseEvent) error {
	// Do nothing - just simulate successful send for testing
	return nil
}

func (m *mockKafkaProducer) Topic() string {
	return "test-topic"
}

// TestSuite holds all integration tests.
type TestSuite struct {
	suite.Suite
	tcSetup        *TestContainersSetup
	httpServer     *server.HTTPServer
	baseURL        string
	productService service.ProductServiceInterface
	ctx            context.Context
}

// SetupSuite runs once before all tests.
func (s *TestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Setup testcontainers
	s.tcSetup = NewTestContainersSetup()
	err := s.tcSetup.SetupPostgres()
	require.NoError(s.T(), err)

	// Setup logger
	appLogger := logger.NewLogrusLogger(4) // Debug level

	// Setup dataStore
	dataStore := repository.NewDataStore(s.tcSetup.DbPool)

	// Setup product service with mock Kafka producers for testing
	mockProducer := &mockKafkaProducer{}
	s.productService = service.NewProductService(
		dataStore,
		mockProducer,
		mockProducer,
		mockProducer,
	)

	// Create test providers with product service already set to bypass Kafka initialization
	testProviders := &provider.Providers{
		DataStore:      dataStore,
		KafkaAdmin:     nil,              // nil for testing
		ProductService: s.productService, // Pre-initialize to avoid Kafka setup
	}

	// Create a test config
	testConfig := &config.Config{
		HTTPServer: &config.HTTPServerConfig{
			Port:        10080 + (int(time.Now().UnixNano()/1000000) % 1000),
			GracePeriod: 5,
		},
		App: &config.AppConfig{
			Environment: "test", // Set to test environment
		},
	}

	// Setup HTTP server
	s.httpServer = server.NewHTTPServer(testConfig, appLogger, testProviders)

	// Start HTTP server in goroutine
	go func() {
		if err := s.httpServer.StartHTTP(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.T().Errorf("HTTP server error: %v", err)
			}
		}
	}()

	// Wait for server to start
	time.Sleep(300 * time.Millisecond)

	s.baseURL = fmt.Sprintf("http://localhost:%d", testConfig.HTTPServer.Port)

	// Verify server is running
	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, s.baseURL+"/health", http.NoBody)
	require.NoError(s.T(), err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)

	if err := resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

// TearDownSuite runs once after all tests.
func (s *TestSuite) TearDownSuite() {
	if s.httpServer != nil {
		s.httpServer.Shutdown()
	}

	if s.tcSetup != nil {
		s.tcSetup.Cleanup()
	}
}

// SetupTest runs before each test.
func (s *TestSuite) SetupTest() {
	// Clean up products table before each test
	err := s.tcSetup.CleanupData()
	require.NoError(s.T(), err)
}

// Helper methods for making HTTP requests.
func (s *TestSuite) makeRequest(
	method string,
	endpoint string,
	body interface{},
) (*http.Response, error) {
	var reqBody []byte

	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(
		s.ctx,
		method,
		s.baseURL+endpoint,
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Add mock authentication headers for testing
	req.Header.Set("X-User-ID", "550e8400-e29b-41d4-a716-446655440000") // Valid UUID format
	req.Header.Set("X-Email", "test@example.com")
	req.Header.Set("X-Roles", "admin,user")
	req.Header.Set("X-IsActive", "true")

	client := &http.Client{Timeout: 10 * time.Second}

	return client.Do(req)
}

func (s *TestSuite) parseResponse(resp *http.Response, target interface{}) error {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.T().Errorf("failed to close response body: %v", err)
		}
	}()

	return json.NewDecoder(resp.Body).Decode(target)
}
