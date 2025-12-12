package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
	"github.com/shopspring/decimal"
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
		resp, err := s.makeRequest("POST", "", product)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, resp.StatusCode)

		if err = resp.Body.Close(); err != nil {
			s.T().Errorf("failed to close response body: %v", err)
		}
	}

	// Test getting all products with default pagination
	resp, err := s.makeRequest("GET", "", nil)
	s.Require().NoError(err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	s.Equal(http.StatusOK, resp.StatusCode)

	var productList dto.WebResponse[[]productDto.ProductResponse, dto.CursorPagination]

	err = s.parseResponse(resp, &productList)
	s.Require().NoError(err)

	s.Equal("success", productList.Message)
	s.NotNil(productList.Pagination)
	s.Len(productList.Data, 3)
}

func (s *ProductListTestSuite) TestGetProductsWithPagination() {
	// Create test products
	for i := 1; i <= 5; i++ {
		product := productDto.CreateProductRequest{
			Name:     fmt.Sprintf("Product %d", i),
			Price:    decimal.NewFromFloat(float64(i * 10)),
			Quantity: int64(i * 5),
		}
		resp, err := s.makeRequest("POST", "", product)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, resp.StatusCode)

		if err = resp.Body.Close(); err != nil {
			s.T().Errorf("failed to close response body: %v", err)
		}
	}

	// First page - get first 2 items
	resp, err := s.makeRequest("GET", "?limit=2", nil)
	s.Require().NoError(err)

	var firstPage dto.WebResponse[[]productDto.ProductResponse, dto.CursorPagination]

	err = s.parseResponse(resp, &firstPage)
	s.Require().NoError(err)

	if err = resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}

	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("success", firstPage.Message)
	s.NotNil(firstPage.Pagination)
	s.Len(firstPage.Data, 2)
	s.True(firstPage.Pagination.HasNext)
	s.NotEmpty(firstPage.Pagination.NextCursor)

	// Second page - use cursor from first page
	nextCursor := firstPage.Pagination.NextCursor
	resp2, err := s.makeRequest("GET", fmt.Sprintf("?limit=2&next_cursor=%s", nextCursor), nil)
	s.Require().NoError(err)

	defer func() {
		if cerr := resp2.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	s.Equal(http.StatusOK, resp2.StatusCode)

	var secondPage dto.WebResponse[[]productDto.ProductResponse, dto.CursorPagination]

	err = s.parseResponse(resp2, &secondPage)
	s.Require().NoError(err)

	s.Equal("success", secondPage.Message)
	s.NotNil(secondPage.Pagination)
	s.Equal(int64(2), secondPage.Pagination.Limit)
	s.Len(secondPage.Data, 2)
}

// TestProductListSuite runs the product listing test suite.
func TestProductListSuite(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(ProductListTestSuite))
}
