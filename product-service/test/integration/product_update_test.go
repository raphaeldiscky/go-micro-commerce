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

	resp, err := s.makeRequest("POST", "", createReq)
	s.Require().NoError(err)

	s.Equal(http.StatusCreated, resp.StatusCode)

	var createResponse dto.WebResponse[productDto.ProductResponse, any]

	err = s.parseResponse(resp, &createResponse)
	s.Require().NoError(err)

	// Update the product
	updateReq := productDto.UpdateProductRequest{
		ID:       createResponse.Data.ID,
		Name:     "Updated Product",
		Price:    decimal.NewFromFloat(35.00),
		Quantity: 30,
	}

	// Close the previous response body before reassigning
	if err = resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}

	resp, err = s.makeRequest(
		"PUT",
		fmt.Sprintf("/%s", createResponse.Data.ID),
		updateReq,
	)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	var updateResponse dto.WebResponse[productDto.ProductResponse, any]

	err = s.parseResponse(resp, &updateResponse)
	s.Require().NoError(err)

	s.Equal(createResponse.Data.ID, updateResponse.Data.ID)
	s.Equal("Updated Product", updateResponse.Data.Name)
	s.True(updateResponse.Data.Price.Equal(decimal.NewFromFloat(35.00)))
	s.Equal(int64(30), updateResponse.Data.Quantity)
	s.True(updateResponse.Data.UpdatedAt.After(createResponse.Data.UpdatedAt))

	// Close the final response body
	if err = resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
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
		fmt.Sprintf("/%s", nonExistentID),
		updateReq,
	)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)

	if err = resp.Body.Close(); err != nil {
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

	resp, err := s.makeRequest("POST", "", createReq)
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
		s.Run(tc.name, func() {
			resp, err = s.makeRequest(
				"PUT",
				fmt.Sprintf("/%s", createResponse.Data.ID),
				tc.request,
			)
			s.Require().NoError(err)
			s.Equal(tc.wantCode, resp.StatusCode)

			defer func() {
				if cerr := resp.Body.Close(); cerr != nil {
					s.T().Errorf("failed to close response body: %v", cerr)
				}
			}()
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
