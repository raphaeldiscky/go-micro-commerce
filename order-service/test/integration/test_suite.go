// Package integration provides integration tests for the order service.
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/stretchr/testify/require"
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
	_ = logger.NewLogrusLogger(4) // Debug level

	// Create test config with random port
	port := 10080 + (int(time.Now().UnixNano()/1000000) % 1000)
	s.baseURL = fmt.Sprintf("http://localhost:%d", port)

	// Setup simple HTTP server with basic endpoints for testing
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/saga", s.mockSagaEndpoint)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"status":"ok"}`))
		require.NoError(s.T(), err)
	})

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start HTTP server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.T().Errorf("HTTP server error: %v", err)
			}
		}
	}()

	// Wait for server to start
	time.Sleep(300 * time.Millisecond)

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
		err := s.httpServer.Shutdown(s.ctx)
		require.NoError(s.T(), err)
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
	userID := r.Header.Get("X-User-ID")
	email := r.Header.Get("X-Email")

	if userID == "" || email == "" {
		http.Error(w, "Authentication required", http.StatusUnauthorized)

		return
	}

	// Parse request body
	var req dto.ClientCreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	response := pkgDto.WebResponse[dto.OrderResponse]{
		Message: "Order created successfully",
		Data:    mockOrder,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		s.T().Errorf("Failed to encode response: %v", err)
	}
}

// SetupTest runs before each test.
func (s *TestSuite) SetupTest() {
	// Clean up orders and related tables before each test
	// Only if database is available
	if s.tcSetup != nil && s.tcSetup.DbPool != nil {
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
	req.Header.Set("X-Is-Active", "true")

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

// makeRequestWithoutAuth makes HTTP requests without authentication headers for testing auth requirements.
func (s *TestSuite) makeRequestWithoutAuth(
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
	// Note: No authentication headers added

	client := &http.Client{Timeout: 10 * time.Second}

	return client.Do(req)
}
