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

// ProductDeleteTestSuite holds product deletion tests.
type ProductDeleteTestSuite struct {
	TestSuite
}

func (s *ProductDeleteTestSuite) TestDeleteProduct() {
	// First create a product
	createReq := productDto.CreateProductRequest{
		Name:     "Product to Delete",
		Price:    decimal.NewFromFloat(15.00),
		Quantity: 10,
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

	err = s.parseResponse(resp, &createResponse)
	s.Require().NoError(err)

	// Delete the product
	resp, err = s.makeRequest(
		"DELETE",
		fmt.Sprintf("/v1/%s", createResponse.Data.ID),
		nil,
	)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	if err = resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}

	// Verify product is deleted
	resp, err = s.makeRequest("GET", fmt.Sprintf("/v1/%s", createResponse.Data.ID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)

	if err = resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

func (s *ProductDeleteTestSuite) TestDeleteProductNotFound() {
	nonExistentID := uuid.New()

	resp, err := s.makeRequest("DELETE", fmt.Sprintf("/v1/%s", nonExistentID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)

	if err = resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

func (s *ProductDeleteTestSuite) TestDeleteProductInvalidID() {
	resp, err := s.makeRequest("DELETE", "/v1/invalid-uuid", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)

	if err = resp.Body.Close(); err != nil {
		s.T().Errorf("failed to close response body: %v", err)
	}
}

// TestProductDeleteSuite runs the product deletion test suite.
func TestProductDeleteSuite(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(ProductDeleteTestSuite))
}
