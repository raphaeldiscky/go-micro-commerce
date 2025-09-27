package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
	"github.com/shopspring/decimal"
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
	s.Require().NoError(err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	s.Equal(http.StatusCreated, resp.StatusCode)

	var createResponse dto.WebResponse[productDto.ProductResponse, any]

	err = s.parseResponse(resp, &createResponse)
	s.Require().NoError(err)

	// Test getting the product
	resp, err = s.makeRequest("GET", fmt.Sprintf("/v1/%s", createResponse.Data.ID), nil)
	s.Require().NoError(err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	s.Equal(http.StatusOK, resp.StatusCode)

	var getResponse dto.WebResponse[productDto.ProductResponse, any]

	err = s.parseResponse(resp, &getResponse)
	s.Require().NoError(err)

	s.Equal(createResponse.Data.ID, getResponse.Data.ID)
	s.Equal("Test Product", getResponse.Data.Name)
	s.True(getResponse.Data.Price.Equal(decimal.NewFromFloat(19.99)))
	s.Equal(int64(50), getResponse.Data.Quantity)
}

func (s *ProductGetTestSuite) TestGetProductNotFound() {
	nonExistentID := uuid.New()

	resp, err := s.makeRequest("GET", fmt.Sprintf("/v1/%s", nonExistentID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)

	if err = resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

func (s *ProductGetTestSuite) TestGetProductInvalidID() {
	resp, err := s.makeRequest("GET", "/v1/invalid-uuid", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)

	if err = resp.Body.Close(); err != nil {
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
