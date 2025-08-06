package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/raphaeldiscky/go-micro-template/product-service/internal/dto"
)

// ProductDeleteTestSuite holds product deletion tests.
type ProductDeleteTestSuite struct {
	IntegrationTestSuite
}

func (s *ProductDeleteTestSuite) TestDeleteProduct() {
	// First create a product
	createReq := dto.CreateProductRequest{
		Name:     "Product to Delete",
		Price:    decimal.NewFromFloat(15.00),
		Quantity: 10,
	}

	resp, err := s.makeRequest("POST", "/api/v1/products", createReq)
	require.NoError(s.T(), err)

	var createdProduct dto.ProductResponse
	err = s.parseResponse(resp, &createdProduct)
	require.NoError(s.T(), err)

	// Close response body after parsing
	if cerr := resp.Body.Close(); cerr != nil {
		require.NoError(s.T(), cerr)
	}

	// Delete the product
	resp, err = s.makeRequest(
		"DELETE",
		fmt.Sprintf("/api/v1/products/%s", createdProduct.ID),
		nil,
	)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), http.StatusNoContent, resp.StatusCode)

	if cerr := resp.Body.Close(); cerr != nil {
		require.NoError(s.T(), cerr)
	}

	// Verify product is deleted
	resp, err = s.makeRequest("GET", fmt.Sprintf("/api/v1/products/%s", createdProduct.ID), nil)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)

	if cerr := resp.Body.Close(); cerr != nil {
		require.NoError(s.T(), cerr)
	}
}

func (s *ProductDeleteTestSuite) TestDeleteProductNotFound() {
	nonExistentID := uuid.New()

	resp, err := s.makeRequest("DELETE", fmt.Sprintf("/api/v1/products/%s", nonExistentID), nil)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), http.StatusInternalServerError, resp.StatusCode)

	if cerr := resp.Body.Close(); cerr != nil {
		require.NoError(s.T(), cerr)
	}
}

func (s *ProductDeleteTestSuite) TestDeleteProductInvalidID() {
	resp, err := s.makeRequest("DELETE", "/api/v1/products/invalid-uuid", nil)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)

	if cerr := resp.Body.Close(); cerr != nil {
		require.NoError(s.T(), cerr)
	}
}

// TestProductDeleteSuite runs the product deletion test suite.
func TestProductDeleteSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(ProductDeleteTestSuite))
}
