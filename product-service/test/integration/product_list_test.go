package integration

// import (
// 	"fmt"
// 	"net/http"
// 	"testing"

// 	"github.com/shopspring/decimal"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"github.com/stretchr/testify/suite"

// 	"github.com/raphaeldiscky/go-micro-template/product-service/internal/dto"
// )

// // ProductListTestSuite holds product listing tests.
// type ProductListTestSuite struct {
// 	IntegrationTestSuite
// }

// func (s *ProductListTestSuite) TestGetProducts() {
// 	// Create test products
// 	products := []dto.CreateProductRequest{
// 		{Name: "Product 1", Price: decimal.NewFromFloat(10.00), Quantity: 5},
// 		{Name: "Product 2", Price: decimal.NewFromFloat(20.00), Quantity: 10},
// 		{Name: "Product 3", Price: decimal.NewFromFloat(30.00), Quantity: 15},
// 	}

// 	for _, product := range products {
// 		resp, err := s.makeRequest("POST", "/api/v1/products", product)
// 		require.NoError(s.T(), err)

// 		// Close response body immediately since we don't need to parse it
// 		if cerr := resp.Body.Close(); cerr != nil {
// 			require.NoError(s.T(), cerr)
// 		}
// 	}

// 	// Test getting all products with default pagination
// 	resp, err := s.makeRequest("GET", "/api/v1/products", nil)
// 	require.NoError(s.T(), err)

// 	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

// 	var productList dto.ProductListResponse
// 	err = s.parseResponse(resp, &productList)
// 	require.NoError(s.T(), err)

// 	// Close response body after parsing
// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	assert.Equal(s.T(), int64(3), productList.Total)
// 	assert.Equal(s.T(), 10, productList.Limit)
// 	assert.Equal(s.T(), 0, productList.Offset)
// 	assert.Len(s.T(), productList.Products, 3)
// }

// func (s *ProductListTestSuite) TestGetProductsWithPagination() {
// 	// Create test products
// 	for i := 1; i <= 5; i++ {
// 		product := dto.CreateProductRequest{
// 			Name:     fmt.Sprintf("Product %d", i),
// 			Price:    decimal.NewFromFloat(float64(i * 10)),
// 			Quantity: i * 5,
// 		}
// 		resp, err := s.makeRequest("POST", "/api/v1/products", product)
// 		require.NoError(s.T(), err)

// 		// Close response body immediately since we don't need to parse it
// 		if cerr := resp.Body.Close(); cerr != nil {
// 			require.NoError(s.T(), cerr)
// 		}
// 	}

// 	// Test pagination
// 	resp, err := s.makeRequest("GET", "/api/v1/products?limit=2&offset=1", nil)
// 	require.NoError(s.T(), err)

// 	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

// 	var productList dto.ProductListResponse
// 	err = s.parseResponse(resp, &productList)
// 	require.NoError(s.T(), err)

// 	// Close response body after parsing
// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	assert.Equal(s.T(), int64(5), productList.Total)
// 	assert.Equal(s.T(), 2, productList.Limit)
// 	assert.Equal(s.T(), 1, productList.Offset)
// 	assert.Len(s.T(), productList.Products, 2)
// }

// // TestProductListSuite runs the product listing test suite.
// func TestProductListSuite(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration tests in short mode")
// 	}

// 	suite.Run(t, new(ProductListTestSuite))
// }
