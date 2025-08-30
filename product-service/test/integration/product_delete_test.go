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

// ProductDeleteTestSuite holds product deletion tests.
type ProductDeleteTestSuite struct {
	TestSuite
}

func (s *ProductDeleteTestSuite) TestDeleteProduct() {
	// First create a product
	createReq := productDto.CreateProductRequest{
		Name:     "Product to Delete",
		Price:    decimal.NewFromFloat(15.00),
		Quantity: 10,
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

	// Delete the product
	resp, err = s.makeRequest(
		"DELETE",
		fmt.Sprintf("/v1/%s", createResponse.Data.ID),
		nil,
	)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	if err := resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}

	// Verify product is deleted
	resp, err = s.makeRequest("GET", fmt.Sprintf("/v1/%s", createResponse.Data.ID), nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)

	if err := resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

func (s *ProductDeleteTestSuite) TestDeleteProductNotFound() {
	nonExistentID := uuid.New()

	resp, err := s.makeRequest("DELETE", fmt.Sprintf("/v1/%s", nonExistentID), nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)

	if err := resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

func (s *ProductDeleteTestSuite) TestDeleteProductInvalidID() {
	resp, err := s.makeRequest("DELETE", "/v1/invalid-uuid", nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)

	if err := resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

// TestProductDeleteSuite runs the product deletion test suite.
func TestProductDeleteSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(ProductDeleteTestSuite))
}
