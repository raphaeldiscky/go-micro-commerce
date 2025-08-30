package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	productDto "github.com/raphaeldiscky/go-micro-commerce/product-service/internal/dto"
)

// ProductListTestSuite holds product listing tests.
type ProductListTestSuite struct {
	TestSuite
}

func (s *ProductListTestSuite) TestGetProducts() {
	// Create test products
	products := []productDto.CreateProductRequest{
		{Name: "Product 1", Price: decimal.NewFromFloat(10.00), Quantity: 5},
		{Name: "Product 2", Price: decimal.NewFromFloat(20.00), Quantity: 10},
		{Name: "Product 3", Price: decimal.NewFromFloat(30.00), Quantity: 15},
	}

	for _, product := range products {
		resp, err := s.makeRequest("POST", "/v1", product)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

		if err := resp.Body.Close(); err != nil {
			s.T().Errorf("failed to close response body: %v", err)
		}
	}

	// Test getting all products with default pagination
	resp, err := s.makeRequest("GET", "/v1", nil)
	require.NoError(s.T(), err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var productList dto.WebResponse[[]productDto.ProductResponse]
	err = s.parseResponse(resp, &productList)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "success", productList.Message)
	assert.NotNil(s.T(), productList.Pagination)
	assert.Equal(s.T(), int64(3), productList.Pagination.TotalItem)
	assert.Len(s.T(), productList.Data, 3)
}

func (s *ProductListTestSuite) TestGetProductsWithPagination() {
	// Create test products
	for i := 1; i <= 5; i++ {
		product := productDto.CreateProductRequest{
			Name:     fmt.Sprintf("Product %d", i),
			Price:    decimal.NewFromFloat(float64(i * 10)),
			Quantity: i * 5,
		}
		resp, err := s.makeRequest("POST", "/v1", product)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

		if err := resp.Body.Close(); err != nil {
			s.T().Errorf("failed to close response body: %v", err)
		}
	}

	// Test pagination - using limit=2&page=2 (second page with 2 items)
	resp, err := s.makeRequest("GET", "/v1?limit=2&page=2", nil)
	require.NoError(s.T(), err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var productList dto.WebResponse[[]productDto.ProductResponse]
	err = s.parseResponse(resp, &productList)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), "success", productList.Message)
	assert.NotNil(s.T(), productList.Pagination)
	assert.Equal(s.T(), int64(5), productList.Pagination.TotalItem)
	assert.Equal(s.T(), int64(2), productList.Pagination.Size)
	assert.Equal(s.T(), int64(2), productList.Pagination.Page)
	assert.Len(s.T(), productList.Data, 2)
}

// TestProductListSuite runs the product listing test suite.
func TestProductListSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(ProductListTestSuite))
}
