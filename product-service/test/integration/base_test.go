// Package integration provides integration tests for the product service.
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	postgresrepo "github.com/raphaeldiscky/go-micro-template/product-service/internal/infra/db/postgres"
	handlers "github.com/raphaeldiscky/go-micro-template/product-service/internal/interface/http/handler"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/interface/http/server"
	services "github.com/raphaeldiscky/go-micro-template/product-service/internal/service"
)

// IntegrationTestSuite holds all integration tests.
type IntegrationTestSuite struct {
	suite.Suite
	tcSetup        *TestContainersSetup
	httpServer     *server.HTTPServer
	baseURL        string
	productService services.ProductServiceInterface
	ctx            context.Context
}

// SetupSuite runs once before all tests.
func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Setup testcontainers
	s.tcSetup = NewTestContainersSetup()
	err := s.tcSetup.SetupPostgres()
	require.NoError(s.T(), err)

	// Setup logger
	appLogger := logger.NewLogrusLogger(4) // Debug level

	// Setup repository and service
	productRepo := postgresrepo.NewProductRepositoryPostgres(s.tcSetup.DbPool)
	s.productService = services.NewProductService(productRepo, nil, appLogger)

	// Setup HTTP handlers and server
	productHandler := handlers.NewProductHandler(s.productService)
	s.httpServer = server.NewHTTPServer(productHandler)

	// Use a different port for integration tests to avoid conflicts
	testPort := "10081"

	// Start HTTP server in goroutine
	go func() {
		if err := s.httpServer.Start(testPort); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.T().Errorf("HTTP server error: %v", err)
			}
		}
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	s.baseURL = "http://localhost:" + testPort

	// Verify server is running
	resp, err := http.NewRequestWithContext(s.ctx, http.MethodGet, s.baseURL+"/health", http.NoBody)
	require.NoError(s.T(), err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp2, err := client.Do(resp)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp2.StatusCode)

	if err := resp2.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

// TearDownSuite runs once after all tests.
func (s *IntegrationTestSuite) TearDownSuite() {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			s.T().Errorf("failed to shutdown HTTP server: %v", err)
		}
	}

	if s.tcSetup != nil {
		s.tcSetup.Cleanup()
	}
}

// SetupTest runs before each test.
func (s *IntegrationTestSuite) SetupTest() {
	// Clean up products table before each test
	err := s.tcSetup.CleanupData()
	require.NoError(s.T(), err)
}

// Helper methods for making HTTP requests.
func (s *IntegrationTestSuite) makeRequest(
	method, endpoint string,
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

	client := &http.Client{Timeout: 10 * time.Second}

	return client.Do(req)
}

func (s *IntegrationTestSuite) parseResponse(resp *http.Response, target interface{}) error {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.T().Errorf("failed to close response body: %v", err)
		}
	}()

	return json.NewDecoder(resp.Body).Decode(target)
}

// TestMain sets up test environment.
func TestMain(m *testing.M) {
	// Set test environment variables
	if err := os.Setenv("APP_ENV", "test"); err != nil {
		panic("failed to set APP_ENV: " + err.Error())
	}

	if err := os.Setenv("LOG_LEVEL", "error"); err != nil {
		panic("failed to set LOG_LEVEL: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Exit with the same code as the test run
	os.Exit(code)
}
