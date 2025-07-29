package rest_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/raphaeldiscky/go-ddd-template/internal/app/command"
	"github.com/raphaeldiscky/go-ddd-template/internal/app/common"
	"github.com/raphaeldiscky/go-ddd-template/internal/app/query"
	entity "github.com/raphaeldiscky/go-ddd-template/internal/domain/entity"
	rest "github.com/raphaeldiscky/go-ddd-template/internal/interface/http"
	"github.com/raphaeldiscky/go-ddd-template/internal/interface/http/dto/request"
	"github.com/raphaeldiscky/go-ddd-template/internal/interface/http/dto/response"
	"github.com/raphaeldiscky/go-ddd-template/internal/mocks"
)

func TestCreateSeller(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockSellerService(ctrl)
	controller := rest.NewSellerController(echo.New(), mockService)

	// Create a seller for testing
	sellerRequest := request.CreateSellerRequest{
		Name:  "TestSeller",
		Email: "test@example.com",
	}

	expectedResult := &command.CreateSellerCommandResult{
		Result: &common.SellerResult{
			ID:    uuid.New(),
			Name:  "TestSeller",
			Email: "test@example.com",
		},
	}

	mockService.EXPECT().CreateSeller(gomock.Any()).Return(expectedResult, nil).Times(1)

	sellerJSON, err := json.Marshal(sellerRequest)
	if err != nil {
		t.Fatalf("Failed to marshal seller request: %s", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sellers", bytes.NewReader(sellerJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	if err := controller.CreateSellerController(c); err != nil {
		t.Fatal(err)
	}

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)

	var createdSeller entity.Seller
	err = json.Unmarshal(rec.Body.Bytes(), &createdSeller)
	assert.NoError(t, err)
	assert.Equal(t, sellerRequest.Name, createdSeller.Name)
	assert.Equal(t, sellerRequest.Email, createdSeller.Email)
}

func TestPutSeller(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockSellerService(ctrl)
	controller := rest.NewSellerController(echo.New(), mockService)

	sellerID := uuid.New()
	createResult := &command.CreateSellerCommandResult{
		Result: &common.SellerResult{
			ID:    sellerID,
			Name:  "TestSeller",
			Email: "test@example.com",
		},
	}

	updateResult := &command.UpdateSellerCommandResult{
		Result: &common.SellerResult{
			ID:    sellerID,
			Name:  "updatedName",
			Email: "test@example.com",
		},
	}

	// First expect the CreateSeller call
	mockService.EXPECT().CreateSeller(gomock.Any()).Return(createResult, nil).Times(1)

	createdSeller, err := mockService.CreateSeller(&command.CreateSellerCommand{
		Name:  "TestSeller",
		Email: "test@example.com",
	})
	assert.NoError(t, err)

	updateRequest := request.UpdateSellerRequest{
		ID:   createdSeller.Result.ID,
		Name: "updatedName",
	}

	// Expect the UpdateSeller call
	mockService.EXPECT().UpdateSeller(gomock.Any()).Return(updateResult, nil).Times(1)

	sellerJSON, err := json.Marshal(updateRequest)
	if err != nil {
		t.Fatalf("Failed to marshal update request: %s", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/sellers", bytes.NewReader(sellerJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	if err := controller.PutSellerController(c); err != nil {
		t.Fatal(err)
	}

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	var receivedResponse response.SellerResponse
	err = json.Unmarshal(rec.Body.Bytes(), &receivedResponse)
	assert.NoError(t, err)

	assert.Equal(t, updateRequest.Name, receivedResponse.Name)
}

func TestDeleteSeller(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockSellerService(ctrl)
	controller := rest.NewSellerController(echo.New(), mockService)

	sellerID := uuid.New()
	createResult := &command.CreateSellerCommandResult{
		Result: &common.SellerResult{
			ID:    sellerID,
			Name:  "TestSeller",
			Email: "test@example.com",
		},
	}

	// First expect the CreateSeller call
	mockService.EXPECT().CreateSeller(gomock.Any()).Return(createResult, nil).Times(1)

	createdSeller, err := mockService.CreateSeller(&command.CreateSellerCommand{
		Name:  "TestSeller",
		Email: "test@example.com",
	})
	assert.NoError(t, err)

	// Expect the DeleteSeller call
	mockService.EXPECT().DeleteSeller(createdSeller.Result.ID).Return(nil).Times(1)

	req := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/api/v1/sellers/%s", createdSeller.Result.ID),
		http.NoBody,
	)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	if err := controller.DeleteSellerController(c); err != nil {
		t.Fatal(err)
	}

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestGetSellerById(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockSellerService(ctrl)
	controller := rest.NewSellerController(echo.New(), mockService)

	sellerID := uuid.New()
	createResult := &command.CreateSellerCommandResult{
		Result: &common.SellerResult{
			ID:    sellerID,
			Name:  "TestSeller",
			Email: "test@example.com",
		},
	}

	findResult := &query.SellerQueryResult{
		Result: &common.SellerResult{
			ID:    sellerID,
			Name:  "TestSeller",
			Email: "test@example.com",
		},
	}

	// First expect the CreateSeller call
	mockService.EXPECT().CreateSeller(gomock.Any()).Return(createResult, nil).Times(1)

	createdSeller, err := mockService.CreateSeller(&command.CreateSellerCommand{
		Name:  "TestSeller",
		Email: "test@example.com",
	})
	assert.NoError(t, err)

	// Expect the FindSellerByID call
	mockService.EXPECT().FindSellerByID(createdSeller.Result.ID).Return(findResult, nil).Times(1)

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/v1/sellers/%s", createdSeller.Result.ID),
		http.NoBody,
	)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	if err := controller.GetSellerByIDController(c); err != nil {
		t.Fatal(err)
	}

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	var fetchedSeller response.SellerResponse
	err = json.Unmarshal(rec.Body.Bytes(), &fetchedSeller)
	assert.NoError(t, err)

	assert.Equal(t, createdSeller.Result.ID.String(), fetchedSeller.ID)
	assert.Equal(t, createdSeller.Result.Name, fetchedSeller.Name)
}

func TestGetAllSellers(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockSellerService(ctrl)
	controller := rest.NewSellerController(echo.New(), mockService)

	seller1Id := uuid.New()
	seller2Id := uuid.New()

	createResult1 := &command.CreateSellerCommandResult{
		Result: &common.SellerResult{
			ID:    seller1Id,
			Name:  "TestSeller1",
			Email: "test1@example.com",
		},
	}

	createResult2 := &command.CreateSellerCommandResult{
		Result: &common.SellerResult{
			ID:    seller2Id,
			Name:  "TestSeller2",
			Email: "test2@example.com",
		},
	}

	findAllResult := &query.SellerQueryListResult{
		Result: []*common.SellerResult{
			{
				ID:    seller1Id,
				Name:  "TestSeller1",
				Email: "test1@example.com",
			},
			{
				ID:    seller2Id,
				Name:  "TestSeller2",
				Email: "test2@example.com",
			},
		},
	}

	// Expect the CreateSeller calls
	mockService.EXPECT().CreateSeller(gomock.Any()).Return(createResult1, nil).Times(1)
	mockService.EXPECT().CreateSeller(gomock.Any()).Return(createResult2, nil).Times(1)

	_, err := mockService.CreateSeller(&command.CreateSellerCommand{
		Name:  "TestSeller1",
		Email: "test1@example.com",
	})
	assert.NoError(t, err)

	_, err = mockService.CreateSeller(&command.CreateSellerCommand{
		Name:  "TestSeller2",
		Email: "test2@example.com",
	})
	assert.NoError(t, err)

	// Expect the FindAllSellers call
	mockService.EXPECT().FindAllSellers().Return(findAllResult, nil).Times(1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sellers", http.NoBody)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Act
	if err := controller.GetAllSellersController(c); err != nil {
		t.Fatal(err)
	}

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	var sellers response.ListSellersResponse
	err = json.Unmarshal(rec.Body.Bytes(), &sellers)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(sellers.Sellers))
}
