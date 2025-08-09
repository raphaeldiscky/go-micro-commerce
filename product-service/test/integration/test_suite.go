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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/handler"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/server"
	"github.com/raphaeldiscky/go-micro-template/product-service/internal/service"
)

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

	// Setup event publisher (optional, can be nil for tests)
	topics := constant.NewProductTopics()
	s.productService = service.NewProductService(dataStore, nil, topics, appLogger)

	// Setup HTTP handlers and server
	productHandler := handler.NewProductHandler(s.productService)
	s.httpServer = server.NewHTTPServer(productHandler, nil, appLogger)

	// Use a unique port for each test suite to avoid conflicts
	// Generate a port number based on current time to make it unique
	basePort := 10080 + (int(time.Now().UnixNano()/1000000) % 1000)
	testPort := fmt.Sprintf("%d", basePort)

	// Start HTTP server in goroutine
	go func() {
		if err := s.httpServer.Start(testPort); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.T().Errorf("HTTP server error: %v", err)
			}
		}
	}()

	// Wait for server to start
	time.Sleep(300 * time.Millisecond)

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
func (s *TestSuite) TearDownSuite() {
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
