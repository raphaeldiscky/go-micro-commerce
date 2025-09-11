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
	require.NoError(s.T(), err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var createResponse dto.WebResponse[productDto.ProductResponse]

	require.NoError(s.T(), s.parseResponse(resp, &createResponse))
	productID := createResponse.Data.ID

	// --- Read ---
	resp, err = s.makeRequest("GET", fmt.Sprintf("/v1/%s", productID), nil)
	require.NoError(s.T(), err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var getResponse dto.WebResponse[productDto.ProductResponse]

	require.NoError(s.T(), s.parseResponse(resp, &getResponse))
	assert.Equal(s.T(), "Workflow Test Product", getResponse.Data.Name)
	assert.True(s.T(), getResponse.Data.Price.Equal(decimal.NewFromFloat(99.99)))
	assert.Equal(s.T(), int64(100), getResponse.Data.Quantity)

	// --- Update ---
	updateReq := productDto.UpdateProductRequest{
		ID:       productID,
		Name:     "Updated Workflow Product",
		Price:    decimal.NewFromFloat(149.99),
		Quantity: 150,
	}

	resp, err = s.makeRequest("PUT", fmt.Sprintf("/v1/%s", productID), updateReq)
	require.NoError(s.T(), err)

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.T().Errorf("failed to close response body: %v", cerr)
		}
	}()

	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	var updateResponse dto.WebResponse[productDto.ProductResponse]

	require.NoError(s.T(), s.parseResponse(resp, &updateResponse))
	assert.Equal(s.T(), "Updated Workflow Product", updateResponse.Data.Name)
	assert.True(s.T(), updateResponse.Data.Price.Equal(decimal.NewFromFloat(149.99)))
	assert.Equal(s.T(), int64(150), updateResponse.Data.Quantity)

	// --- Delete ---
	resp, err = s.makeRequest("DELETE", fmt.Sprintf("/v1/%s", productID), nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

	// --- Confirm Deletion ---
	resp, err = s.makeRequest("GET", fmt.Sprintf("/v1/%s", productID), nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)

	if err := resp.Body.Close(); err != nil {
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
