package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	productDto "github.com/raphaeldiscky/go-micro-commerce/product-service/internal/dto"
)

// ProductGetTestSuite holds product retrieval tests.
type ProductGetTestSuite struct {
	TestSuite
}

func (s *ProductGetTestSuite) TestGetProduct() {
	// First create a product
	createReq := productDto.CreateProductRequest{
		Name:     "Test Product",
		Price:    decimal.NewFromFloat(19.99),
		Quantity: 50,
	}

	resp, err := s.makeRequest("POST", "/v1", createReq)
	require.NoError(s.T(), err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var createResponse dto.WebResponse[productDto.ProductResponse]

	err = s.parseResponse(resp, &createResponse)
	require.NoError(s.T(), err)

	// Test getting the product
	resp, err = s.makeRequest("GET", fmt.Sprintf("/v1/%s", createResponse.Data.ID), nil)
	require.NoError(s.T(), err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var getResponse dto.WebResponse[productDto.ProductResponse]

	err = s.parseResponse(resp, &getResponse)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), createResponse.Data.ID, getResponse.Data.ID)
	assert.Equal(s.T(), "Test Product", getResponse.Data.Name)
	assert.True(s.T(), getResponse.Data.Price.Equal(decimal.NewFromFloat(19.99)))
	assert.Equal(s.T(), int64(50), getResponse.Data.Quantity)
}

func (s *ProductGetTestSuite) TestGetProductNotFound() {
	nonExistentID := uuid.New()

	resp, err := s.makeRequest("GET", fmt.Sprintf("/v1/%s", nonExistentID), nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)

	if err := resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

func (s *ProductGetTestSuite) TestGetProductInvalidID() {
	resp, err := s.makeRequest("GET", "/v1/invalid-uuid", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)

	if err := resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

// TestProductGetSuite runs the product retrieval test suite.
func TestProductGetSuite(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(ProductGetTestSuite))
}
