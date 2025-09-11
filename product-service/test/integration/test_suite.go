// Package integration_test provides integration tests for the product service.
package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/event"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/testutils"
	"github.com/stretchr/testify/suite"
	"golang.org/x/time/rate"

	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/provider"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/server"
	"github.com/raphaeldiscky/go-micro-commerce/product-service/internal/service"
)

// mockKafkaProducer is a mock implementation of KafkaProducerInterface for testing.
type mockKafkaProducer struct{}

func (m *mockKafkaProducer) Send(_ context.Context, _ event.BaseEvent) error {
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

const (
	logLevelDebug       = 4
	serverStartupDelay  = 300 * time.Millisecond
	healthCheckTimeout  = 5 * time.Second
	httpRequestTimeout  = 10 * time.Second
	rateLimiterForTests = 10000
)

// SetupSuite runs once before all tests.
func (s *TestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Setup testcontainers
	s.tcSetup = NewTestContainersSetup()
	err := s.tcSetup.SetupPostgres()
	s.Require().NoError(err)

	// Setup logger
	appLogger := logger.NewLogrusLogger(logLevelDebug) // Debug level

	// Setup dataStore with nil Redis client for testing (cache will be bypassed)
	dataStore := repository.NewDataStore(s.tcSetup.DBPool, nil)

	// Setup product service with mock Kafka producers for testing
	mockProducer := &mockKafkaProducer{}
	s.productService = service.NewProductService(
		dataStore,
		nil,
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

	port, err := testutils.GetFreePort()
	s.Require().NoError(err)

	// Create a test config
	testConfig := &config.Config{
		HTTPServer: &config.HTTPServerConfig{
			Host:        "localhost",
			Port:        port,
			RateLimiter: rate.Limit(rateLimiterForTests), // High limit for tests
		},
		App: &config.AppConfig{
			Environment: "test", // Set to test environment
		},
	}

	// Setup HTTP server
	s.httpServer = server.NewHTTPServer(s.ctx, testConfig, appLogger, testProviders)

	// Start HTTP server in goroutine
	go func() {
		if err = s.httpServer.Start(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.T().Errorf("HTTP server error: %v", err)
			}
		}
	}()

	// Wait for server to start
	time.Sleep(serverStartupDelay)

	s.baseURL = fmt.Sprintf("http://localhost:%d", testConfig.HTTPServer.Port)

	// Verify server is running
	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, s.baseURL+"/health", http.NoBody)
	s.Require().NoError(err)

	client := &http.Client{Timeout: healthCheckTimeout}
	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	if err = resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

// TearDownSuite runs once after all tests.
func (s *TestSuite) TearDownSuite() {
	if s.httpServer != nil {
		err := s.httpServer.Shutdown(s.ctx)
		s.Require().NoError(err)
	}

	if s.tcSetup != nil {
		s.tcSetup.Cleanup()
	}
}

// SetupTest runs before each test.
func (s *TestSuite) SetupTest() {
	// Clean up products table before each test
	err := s.tcSetup.CleanupData()
	s.Require().NoError(err)
}

// Helper methods for making HTTP requests.
func (s *TestSuite) makeRequest(
	method string,
	endpoint string,
	body any,
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
	req.Header.Set("X-User-Id", "550e8400-e29b-41d4-a716-446655440000") // Valid UUID format
	req.Header.Set("X-Email", "test@example.com")
	req.Header.Set("X-Roles", "admin,user")
	req.Header.Set("X-Is-Active", "true")

	client := &http.Client{Timeout: httpRequestTimeout}

	return client.Do(req)
}

func (s *TestSuite) parseResponse(resp *http.Response, target any) error {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.T().Errorf("failed to close response body: %v", err)
		}
	}()

	return json.NewDecoder(resp.Body).Decode(target)
}
