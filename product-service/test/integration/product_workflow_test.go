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

// ProductWorkflowTestSuite holds the test suite.
type ProductWorkflowTestSuite struct {
	TestSuite
}

func (s *ProductWorkflowTestSuite) TestCRUDWorkflow() {
	// --- Create ---
	createReq := productDto.CreateProductRequest{
		Name:     "Workflow Test Product",
		Price:    decimal.NewFromFloat(99.99),
		Quantity: 100,
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

	s.Require().NoError(s.parseResponse(resp, &createResponse))
	productID := createResponse.Data.ID

	// --- Read ---
	resp, err = s.makeRequest("GET", fmt.Sprintf("/v1/%s", productID), nil)
	s.Require().NoError(err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	s.Equal(http.StatusOK, resp.StatusCode)

	var getResponse dto.WebResponse[productDto.ProductResponse, any]

	s.Require().NoError(s.parseResponse(resp, &getResponse))
	s.Equal("Workflow Test Product", getResponse.Data.Name)
	s.True(getResponse.Data.Price.Equal(decimal.NewFromFloat(99.99)))
	s.Equal(int64(100), getResponse.Data.Quantity)

	// --- Update ---
	updateReq := productDto.UpdateProductRequest{
		ID:       productID,
		Name:     "Updated Workflow Product",
		Price:    decimal.NewFromFloat(149.99),
		Quantity: 150,
	}

	resp, err = s.makeRequest("PUT", fmt.Sprintf("/v1/%s", productID), updateReq)
	s.Require().NoError(err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	s.Equal(http.StatusOK, resp.StatusCode)

	var updateResponse dto.WebResponse[productDto.ProductResponse, any]

	s.Require().NoError(s.parseResponse(resp, &updateResponse))
	s.Equal("Updated Workflow Product", updateResponse.Data.Name)
	s.True(updateResponse.Data.Price.Equal(decimal.NewFromFloat(149.99)))
	s.Equal(int64(150), updateResponse.Data.Quantity)

	// --- Delete ---
	resp, err = s.makeRequest("DELETE", fmt.Sprintf("/v1/%s", productID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	// --- Confirm Deletion ---
	resp, err = s.makeRequest("GET", fmt.Sprintf("/v1/%s", productID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)

	if err = resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

// Entrypoint to run the test suite.
func TestProductWorkflowSuite(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(ProductWorkflowTestSuite))
}
