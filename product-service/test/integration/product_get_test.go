package integration

// // ProductGetTestSuite holds product retrieval tests.
// type ProductGetTestSuite struct {
// 	TestSuite
// }

// func (s *ProductGetTestSuite) TestGetProduct() {
// 	// First create a product
// 	createReq := dto.CreateProductRequest{
// 		Name:     "Test Product",
// 		Price:    decimal.NewFromFloat(19.99),
// 		Quantity: 50,
// 	}

// 	resp, err := s.makeRequest("POST", "/v1", createReq)
// 	s.NoError(err)

// 	var createdProduct dto.ProductResponse
// 	err = s.parseResponse(resp, &createdProduct)
// 	s.NoError(err)

// 	// Close response body after parsing
// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	// Test getting the product
// 	resp, err = s.makeRequest("GET", fmt.Sprintf("/v1/%s", createdProduct.ID), nil)
// 	s.NoError(err)

// 	assert.Equal(s.T(), http.StatusOK, resp.StatusCode)

// 	var product dto.ProductResponse
// 	err = s.parseResponse(resp, &product)
// 	s.NoError(err)

// 	// Close response body after parsing
// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}

// 	assert.Equal(s.T(), createdProduct.ID, product.ID)
// 	assert.Equal(s.T(), "Test Product", product.Name)
// 	assert.True(s.T(), product.Price.Equal(decimal.NewFromFloat(19.99)))
// 	assert.Equal(s.T(), 50, product.Quantity)
// }

// func (s *ProductGetTestSuite) TestGetProductNotFound() {
// 	nonExistentID := uuid.New()

// 	resp, err := s.makeRequest("GET", fmt.Sprintf("/v1/%s", nonExistentID), nil)
// 	s.NoError(err)

// 	assert.Equal(s.T(), http.StatusNotFound, resp.StatusCode)

// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}
// }

// func (s *ProductGetTestSuite) TestGetProductInvalidID() {
// 	resp, err := s.makeRequest("GET", "/v1/invalid-uuid", nil)
// 	s.NoError(err)

// 	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)

// 	if cerr := resp.Body.Close(); cerr != nil {
// 		require.NoError(s.T(), cerr)
// 	}
// }

// // TestProductGetSuite runs the product retrieval test suite.
// func TestProductGetSuite(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration tests in short mode")
// 	}

// 	suite.Run(t, new(ProductGetTestSuite))
// }
