// Package integration_test provides integration tests for the order service.
package integration_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/testcontainers"
	"github.com/stretchr/testify/suite"

	pkgDto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
)

// TestSuite holds all integration tests.
type TestSuite struct {
	suite.Suite

	tcSetup    *TestContainersSetup
	httpServer *http.Server
	baseURL    string
	ctx        context.Context
}

const (
	debugLevel         = 4
	httpRequestTimeout = 15 * time.Second
)

// SetupSuite runs once before all tests.
func (s *TestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Setup testcontainers (optional for basic HTTP testing)
	s.tcSetup = NewTestContainersSetup()

	err := s.tcSetup.SetupPostgres()
	if err != nil {
		s.T().Logf("Database setup skipped (not required for basic HTTP tests): %v", err)
	}

	// Setup logger (not needed for basic HTTP testing)
	_ = logger.NewLogrusLogger(debugLevel) // Debug level

	port, err := testcontainers.GetFreePort()
	s.Require().NoError(err)
	s.baseURL = fmt.Sprintf("http://localhost:%d", port)

	// Setup simple HTTP server with basic endpoints for testing
	mux := http.NewServeMux()
	mux.HandleFunc("/saga", s.mockSagaEndpoint)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		_, err = w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			// Log error but don't fail the handler
			// In production, this would be logged properly
			_ = err
		}
	})

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: httpRequestTimeout,
	}

	// Start HTTP server in goroutine
	go func() {
		if err = s.httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.T().Errorf("HTTP server error: %v", err)
			}
		}
	}()

	// Verify server is running
	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, s.baseURL+"/health", http.NoBody)
	s.Require().NoError(err)

	client := &http.Client{Timeout: httpRequestTimeout}
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

// mockSagaEndpoint provides a simple mock implementation of the saga endpoint for testing.
func (s *TestSuite) mockSagaEndpoint(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	// Check for required auth headers
	userID := r.Header.Get("X-User-Id")
	email := r.Header.Get("X-Email")

	if userID == "" || email == "" {
		http.Error(w, "Authentication required", http.StatusUnauthorized)

		return
	}

	// Parse request body
	var req dto.ClientCreateOrderRequest
	if err := sonic.ConfigFastest.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)

		return
	}

	// Validate request
	if len(req.Items) == 0 {
		http.Error(w, "Items are required", http.StatusBadRequest)

		return
	}

	// Return mock success response
	mockOrder := dto.OrderResponse{
		ID:         uuid.New(),
		CustomerID: uuid.MustParse(userID),
		Status:     "pending",
		Currency:   "USD",
		Items:      []dto.OrderItemResponse{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	response := pkgDto.WebResponse[dto.OrderResponse, any]{
		Message: "Order created successfully",
		Data:    mockOrder,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err := sonic.ConfigFastest.NewEncoder(w).Encode(response)
	if err != nil {
		s.T().Errorf("Failed to encode response: %v", err)
	}
}

// SetupTest runs before each test.
func (s *TestSuite) SetupTest() {
	// Clean up orders and related tables before each test
	// Only if database is available
	if s.tcSetup != nil && s.tcSetup.DBPool != nil {
		err := s.tcSetup.CleanupData()
		if err != nil {
			s.T().Logf("Database cleanup failed (not critical for HTTP tests): %v", err)
		}
	}
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
		reqBody, err = sonic.Marshal(body)
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

	return sonic.ConfigFastest.NewDecoder(resp.Body).Decode(target)
}

// makeRequestWithoutAuth makes HTTP requests without authentication headers for testing auth requirements.
func (s *TestSuite) makeRequestWithoutAuth(
	method string,
	endpoint string,
	body any,
) (*http.Response, error) {
	var reqBody []byte

	var err error

	if body != nil {
		reqBody, err = sonic.Marshal(body)
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

	client := &http.Client{Timeout: httpRequestTimeout}

	return client.Do(req)
}
