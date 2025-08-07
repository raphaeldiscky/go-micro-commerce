package integration

// import (
// 	"fmt"
// 	"net/http"
// 	"testing"

// 	"github.com/google/uuid"
// 	"github.com/shopspring/decimal"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"github.com/stretchr/testify/suite"

// 	"github.com/raphaeldiscky/go-micro-template/product-service/internal/dto"
// )

// // ProductUpdateTestSuite holds product update tests.
// type ProductUpdateTestSuite struct {
// 	TestSuite
// }

// func (s *ProductUpdateTestSuite) TestUpdateProduct() {
// 	// First create a product
// 	createReq := dto.CreateProductRequest{
// 		Name:     "Original Product",
// 		Price:    decimal.NewFromFloat(25.00),
// 		Quantity: 20,
// 	}

// 	resp, err := s.makeRequest("POST", "/api/v1/products", createReq)
// 		s.NoError(err)

// 	var createdProduct dto.ProductResponse
// 	err = s.parseResponse(resp, &createdProduct)
// 		s.NoError(err)

// 	// Close response body after parsing
// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	// Update the product
// 	updateReq := dto.UpdateProductRequest{
// 		ID:       createdProduct.ID,
// 		Name:     "Updated Product",
// 		Price:    decimal.NewFromFloat(35.00),
// 		Quantity: 30,
// 	}

// 	resp, err = s.makeRequest(
// 		"PUT",
// 		fmt.Sprintf("/api/v1/products/%s", createdProduct.ID),
// 		updateReq,
// 	)
// 		s.NoError(err)

// 	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

// 	var updatedProduct dto.ProductResponse
// 	err = s.parseResponse(resp, &updatedProduct)
// 		s.NoError(err)

// 	// Close response body after parsing
// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	assert.Equal(s.T(), createdProduct.ID, updatedProduct.ID)
// 	assert.Equal(s.T(), "Updated Product", updatedProduct.Name)
// 	assert.True(s.T(), updatedProduct.Price.Equal(decimal.NewFromFloat(35.00)))
// 	assert.Equal(s.T(), 30, updatedProduct.Quantity)
// 	assert.True(s.T(), updatedProduct.UpdatedAt.After(createdProduct.UpdatedAt))
// }

// func (s *ProductUpdateTestSuite) TestUpdateProductNotFound() {
// 	nonExistentID := uuid.New()
// 	updateReq := dto.UpdateProductRequest{
// 		ID:       nonExistentID,
// 		Name:     "Updated Product",
// 		Price:    decimal.NewFromFloat(35.00),
// 		Quantity: 30,
// 	}

// 	resp, err := s.makeRequest(
// 		"PUT",
// 		fmt.Sprintf("/api/v1/products/%s", nonExistentID),
// 		updateReq,
// 	)
// 		s.NoError(err)

// 	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)

// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}
// }

// func (s *ProductUpdateTestSuite) TestUpdateProductValidation() {
// 	// First create a product
// 	createReq := dto.CreateProductRequest{
// 		Name:     "Test Product",
// 		Price:    decimal.NewFromFloat(25.00),
// 		Quantity: 20,
// 	}

// 	resp, err := s.makeRequest("POST", "/api/v1/products", createReq)
// 		s.NoError(err)

// 	var createdProduct dto.ProductResponse
// 	err = s.parseResponse(resp, &createdProduct)
// 		s.NoError(err)

// 	// Close response body after parsing
// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	// Test validation errors
// 	testCases := []struct {
// 		name     string
// 		request  dto.UpdateProductRequest
// 		wantCode int
// 	}{
// 		{
// 			name: "empty name",
// 			request: dto.UpdateProductRequest{
// 				ID:       createdProduct.ID,
// 				Name:     "",
// 				Price:    decimal.NewFromFloat(10.00),
// 				Quantity: 5,
// 			},
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name: "zero price",
// 			request: dto.UpdateProductRequest{
// 				ID:       createdProduct.ID,
// 				Name:     "Updated Product",
// 				Price:    decimal.Zero,
// 				Quantity: 5,
// 			},
// 			wantCode: http.StatusBadRequest,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		s.T().Run(tc.name, func(t *testing.T) {
// 			resp, err := s.makeRequest(
// 				"PUT",
// 				fmt.Sprintf("/api/v1/products/%s", createdProduct.ID),
// 				tc.request,
// 			)
// 			require.NoError(t, err)

// 			assert.Equal(t, tc.wantCode, resp.StatusCode)

// 			if cerr := resp.Body.Close(); cerr != nil {
// 				require.NoError(s.T(), cerr)
// 			}
// 		})
// 	}
// }

// // TestProductUpdateSuite runs the product update test suite.
// func TestProductUpdateSuite(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration tests in short mode")
// 	}

// 	suite.Run(t, new(ProductUpdateTestSuite))
// }
