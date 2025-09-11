package integration_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
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
	s.Require().NoError(err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	// Should return 201 Created for successful order creation
	s.Equal(http.StatusCreated, resp.StatusCode)

	var response pkgDto.WebResponse[dto.OrderResponse]

	s.Require().NoError(s.parseResponse(resp, &response))

	// Basic assertions on response structure
	s.NotNil(response)
	s.NotNil(response.Data)
	s.NotEmpty(response.Data.ID)
	s.Equal("pending", string(response.Data.Status))
	s.Equal("USD", response.Data.Currency)
}

// TestCreateOrderWithSagaInvalidRequest tests saga endpoint with invalid data.
func (s *OrderSagaTestSuite) TestCreateOrderWithSagaInvalidRequest() {
	// Create invalid request (empty items)
	createReq := dto.ClientCreateOrderRequest{
		Items: []dto.CreateOrderItemRequest{}, // Empty items should fail validation
	}

	// Make request to saga endpoint
	resp, err := s.makeRequest("POST", "/v1/saga", createReq)
	s.Require().NoError(err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	// Should return validation error (400 Bad Request)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
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
	s.Require().NoError(err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	// Should return 401 Unauthorized
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

// Entrypoint to run the test suite.
func TestOrderSagaSuite(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(OrderSagaTestSuite))
}
