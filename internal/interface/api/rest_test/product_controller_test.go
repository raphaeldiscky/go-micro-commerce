package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-ddd-template/internal/application/command"
	"github.com/raphaeldiscky/go-ddd-template/internal/application/common"
	"github.com/raphaeldiscky/go-ddd-template/internal/application/query"
	"github.com/raphaeldiscky/go-ddd-template/internal/domain/entities"
	"github.com/raphaeldiscky/go-ddd-template/internal/interface/api/rest/dto/response"
	"github.com/raphaeldiscky/go-ddd-template/internal/mocks"
	"go.uber.org/mock/gomock"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-ddd-template/internal/interface/api/rest"
	"github.com/stretchr/testify/assert"
)

func TestCreateProduct(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	e := echo.New()
	mockService := mocks.NewMockProductService(ctrl)
	reqBody := map[string]interface{}{"Name": "TestProduct", "Price": 9.99, "SellerId": "123e4567-e89b-12d3-a456-426614174000"}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(reqBodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	restCtrl := rest.NewProductController(e, mockService)

	createProductCommandResult := &command.CreateProductCommandResult{
		Result: &common.ProductResult{
			Id:    uuid.New(),
			Name:  "TestProduct",
			Price: 9.99,
		},
	}

	mockService.EXPECT().CreateProduct(gomock.Any()).Return(createProductCommandResult, nil).Times(1)

	// Execute
	err := restCtrl.CreateProductController(c)
	assert.NoError(t, err)

	// Deserialize the response body
	var responseBody map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Remove fields from responseBody that are not present in reqBody
	// For example, remove Id and Seller fields
	delete(responseBody, "Id")
	delete(responseBody, "Seller")
	delete(reqBody, "SellerId")
	delete(responseBody, "CreatedAt")
	delete(responseBody, "UpdatedAt")

	// Assertions
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, reqBody, responseBody)
}

func TestGetAllProducts(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	e := echo.New()
	mockService := mocks.NewMockProductService(ctrl)

	expectedProducts := []*entities.Product{
		{
			Id:    uuid.New(),
			Name:  "TestProduct1",
			Price: 9.99,
		}, {
			Id:    uuid.New(),
			Name:  "TestProduct2",
			Price: 14.99,
		},
	}

	expectedResult := &query.ProductQueryListResult{
		Result: []*common.ProductResult{
			{
				Id:    expectedProducts[0].Id,
				Name:  expectedProducts[0].Name,
				Price: expectedProducts[0].Price,
			},
			{
				Id:    expectedProducts[1].Id,
				Name:  expectedProducts[1].Name,
				Price: expectedProducts[1].Price,
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	restCtrl := rest.NewProductController(e, mockService)
	mockService.EXPECT().FindAllProducts().Return(expectedResult, nil).Times(1)

	var expectedListResponse response.ListProductsResponse
	for _, product := range expectedProducts {
		expectedListResponse.Products = append(expectedListResponse.Products,
			&response.ProductResponse{
				Id:    product.Id.String(),
				Name:  product.Name,
				Price: product.Price,
			})
	}

	// Assertions
	if assert.NoError(t, restCtrl.GetAllProductsController(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		var receivedListResponse response.ListProductsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &receivedListResponse)
		if assert.NoError(t, err) {
			assert.ElementsMatch(t, expectedListResponse.Products, receivedListResponse.Products)
		}
	}
}
