package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pkgDto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/dto"
)

// OrderSagaTestSuite holds the test suite for order saga operations.
type OrderSagaTestSuite struct {
	TestSuite
}

// TestCreateOrderWithSaga tests the saga-based order creation endpoint.
func (s *OrderSagaTestSuite) TestCreateOrderWithSaga() {
	// Create test request
	createReq := dto.ClientCreateOrderRequest{
		Items: []dto.CreateOrderItemRequest{
			{
				ProductID: uuid.New(), // Random product ID for testing
				Quantity:  2,
			},
			{
				ProductID: uuid.New(), // Another random product ID
				Quantity:  1,
			},
		},
	}

	// Make request to saga endpoint
	resp, err := s.makeRequest("POST", "/v1/saga", createReq)
	require.NoError(s.T(), err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	// Should return 201 Created for successful order creation
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var response pkgDto.WebResponse[dto.OrderResponse]

	require.NoError(s.T(), s.parseResponse(resp, &response))

	// Basic assertions on response structure
	assert.NotNil(s.T(), response)
	assert.NotNil(s.T(), response.Data)
	assert.NotEmpty(s.T(), response.Data.ID)
	assert.Equal(s.T(), "pending", string(response.Data.Status))
	assert.Equal(s.T(), "USD", response.Data.Currency)
}

// TestCreateOrderWithSagaInvalidRequest tests saga endpoint with invalid data.
func (s *OrderSagaTestSuite) TestCreateOrderWithSagaInvalidRequest() {
	// Create invalid request (empty items)
	createReq := dto.ClientCreateOrderRequest{
		Items: []dto.CreateOrderItemRequest{}, // Empty items should fail validation
	}

	// Make request to saga endpoint
	resp, err := s.makeRequest("POST", "/v1/saga", createReq)
	require.NoError(s.T(), err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	// Should return validation error (400 Bad Request)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)
}

// TestCreateOrderWithSagaAuthentication tests that authentication is required.
func (s *OrderSagaTestSuite) TestCreateOrderWithSagaAuthentication() {
	// Create test request
	createReq := dto.ClientCreateOrderRequest{
		Items: []dto.CreateOrderItemRequest{
			{
				ProductID: uuid.New(),
				Quantity:  1,
			},
		},
	}

	// Make request without authentication headers
	resp, err := s.makeRequestWithoutAuth("POST", "/v1/saga", createReq)
	require.NoError(s.T(), err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	// Should return 401 Unauthorized
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)
}

// Entrypoint to run the test suite.
func TestOrderSagaSuite(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(OrderSagaTestSuite))
}
