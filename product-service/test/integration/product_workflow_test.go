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

// // ProductWorkflowTestSuite holds the test suite.
// type ProductWorkflowTestSuite struct {
// 	TestSuite
// }

// func (s *ProductWorkflowTestSuite) TestCRUDWorkflow() {
// 	// --- Create ---
// 	createReq := dto.CreateProductRequest{
// 		Name:     "Workflow Test Product",
// 		Price:    decimal.NewFromFloat(99.99),
// 		Quantity: 100,
// 	}

// 	resp, err := s.makeRequest("POST", "/api/v1/products", createReq)
// 	require.NoError(s.T(), err)

// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	var product dto.ProductResponse

// 	require.NoError(s.T(), s.parseResponse(resp, &product))
// 	productID := product.ID

// 	// --- Read ---
// 	resp, err = s.makeRequest("GET", fmt.Sprintf("/api/v1/products/%s", productID), nil)
// 	require.NoError(s.T(), err)

// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
// 	require.NoError(s.T(), s.parseResponse(resp, &product))
// 	assert.Equal(s.T(), "Workflow Test Product", product.Name)

// 	// --- Update ---
// 	updateReq := dto.UpdateProductRequest{
// 		ID:       productID,
// 		Name:     "Updated Workflow Product",
// 		Price:    decimal.NewFromFloat(149.99),
// 		Quantity: 150,
// 	}

// 	resp, err = s.makeRequest("PUT", fmt.Sprintf("/api/v1/products/%s", productID), updateReq)
// 	require.NoError(s.T(), err)

// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)
// 	require.NoError(s.T(), s.parseResponse(resp, &product))
// 	assert.Equal(s.T(), "Updated Workflow Product", product.Name)
// 	assert.True(s.T(), product.Price.Equal(decimal.NewFromFloat(149.99)))

// 	// --- Delete ---
// 	resp, err = s.makeRequest("DELETE", fmt.Sprintf("/api/v1/products/%s", productID), nil)
// 	require.NoError(s.T(), err)

// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	assert.Equal(s.T(), http.StatusNoContent, resp.StatusCode)

// 	// --- Confirm Deletion ---
// 	resp, err = s.makeRequest("GET", fmt.Sprintf("/api/v1/products/%s", productID), nil)
// 	require.NoError(s.T(), err)

// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)
// }

// // Entrypoint to run the test suite.
// func TestProductWorkflowSuite(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration tests in short mode")
// 	}

// 	suite.Run(t, new(ProductWorkflowTestSuite))
// }
