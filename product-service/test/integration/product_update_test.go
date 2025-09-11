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

// ProductUpdateTestSuite holds product update tests.
type ProductUpdateTestSuite struct {
	TestSuite
}

func (s *ProductUpdateTestSuite) TestUpdateProduct() {
	// First create a product
	createReq := productDto.CreateProductRequest{
		Name:     "Original Product",
		Price:    decimal.NewFromFloat(25.00),
		Quantity: 20,
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

	// Update the product
	updateReq := productDto.UpdateProductRequest{
		ID:       createResponse.Data.ID,
		Name:     "Updated Product",
		Price:    decimal.NewFromFloat(35.00),
		Quantity: 30,
	}

	resp, err = s.makeRequest(
		"PUT",
		fmt.Sprintf("/v1/%s", createResponse.Data.ID),
		updateReq,
	)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var updateResponse dto.WebResponse[productDto.ProductResponse]

	err = s.parseResponse(resp, &updateResponse)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), createResponse.Data.ID, updateResponse.Data.ID)
	assert.Equal(s.T(), "Updated Product", updateResponse.Data.Name)
	assert.True(s.T(), updateResponse.Data.Price.Equal(decimal.NewFromFloat(35.00)))
	assert.Equal(s.T(), int64(30), updateResponse.Data.Quantity)
	assert.True(s.T(), updateResponse.Data.UpdatedAt.After(createResponse.Data.UpdatedAt))
}

func (s *ProductUpdateTestSuite) TestUpdateProductNotFound() {
	nonExistentID := uuid.New()
	updateReq := productDto.UpdateProductRequest{
		ID:       nonExistentID,
		Name:     "Updated Product",
		Price:    decimal.NewFromFloat(35.00),
		Quantity: 30,
	}

	resp, err := s.makeRequest(
		"PUT",
		fmt.Sprintf("/v1/%s", nonExistentID),
		updateReq,
	)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)

	if err := resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

func (s *ProductUpdateTestSuite) TestUpdateProductValidation() {
	// First create a product
	createReq := productDto.CreateProductRequest{
		Name:     "Test Product",
		Price:    decimal.NewFromFloat(25.00),
		Quantity: 20,
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

	// Test validation errors
	testCases := []struct {
		name     string
		request  productDto.UpdateProductRequest
		wantCode int
	}{
		{
			name: "empty name",
			request: productDto.UpdateProductRequest{
				ID:       createResponse.Data.ID,
				Name:     "",
				Price:    decimal.NewFromFloat(10.00),
				Quantity: 5,
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "zero price",
			request: productDto.UpdateProductRequest{
				ID:       createResponse.Data.ID,
				Name:     "Updated Product",
				Price:    decimal.Zero,
				Quantity: 5,
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			resp, err := s.makeRequest(
				"PUT",
				fmt.Sprintf("/v1/%s", createResponse.Data.ID),
				tc.request,
			)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, resp.StatusCode)

			if err := resp.Body.Close(); err != nil {
				t.Errorf("failed to close response body: %v", err)
			}
		})
	}
}

// TestProductUpdateSuite runs the product update test suite.
func TestProductUpdateSuite(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(ProductUpdateTestSuite))
}
