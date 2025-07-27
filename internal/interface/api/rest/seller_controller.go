package rest

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-ddd-template/internal/application/interfaces"
	"github.com/raphaeldiscky/go-ddd-template/internal/interface/api/rest/dto/mapper"
	"github.com/raphaeldiscky/go-ddd-template/internal/interface/api/rest/dto/request"
)

// SellerController handles HTTP requests related to sellers.
type SellerController struct {
	service interfaces.SellerService
}

// NewSellerController initializes a new SellerController.
func NewSellerController(e *echo.Echo, service interfaces.SellerService) *SellerController {
	controller := &SellerController{
		service: service,
	}

	e.POST("/api/v1/sellers", controller.CreateSellerController)
	e.GET("/api/v1/sellers", controller.GetAllSellersController)
	e.GET("/api/v1/sellers/:id", controller.GetSellerByIDController)
	e.PUT("/api/v1/sellers", controller.PutSellerController)
	e.DELETE("/api/v1/sellers/:id", controller.DeleteSellerController)

	return controller
}

// CreateSellerController creates a new seller.
func (sc *SellerController) CreateSellerController(c echo.Context) error {
	var createSellerRequest request.CreateSellerRequest

	if err := c.Bind(&createSellerRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to parse request body",
		})
	}

	sellerCommand, err := createSellerRequest.ToCreateSellerCommand()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid seller Id format",
		})
	}

	commandResult, err := sc.service.CreateSeller(sellerCommand)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create seller",
		})
	}

	response := mapper.ToSellerResponse(commandResult.Result)

	return c.JSON(http.StatusCreated, response)
}

// GetAllSellersController retrieves all sellers.
func (sc *SellerController) GetAllSellersController(c echo.Context) error {
	sellers, err := sc.service.FindAllSellers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch sellers",
		})
	}

	response := mapper.ToSellerListResponse(sellers.Result)

	return c.JSON(http.StatusOK, response)
}

// GetSellerByIDController retrieves a seller by ID.
func (sc *SellerController) GetSellerByIDController(c echo.Context) error {
	// Hack: split the Id from the URL
	// For some reason c.Param("id") doesn't work here
	idRaw := c.Request().URL.Path[len("/api/v1/sellers/"):]

	id, err := uuid.Parse(idRaw)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid seller Id format",
		})
	}

	seller, err := sc.service.FindSellerByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch seller",
		})
	}

	if seller == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Seller not found",
		})
	}

	response := mapper.ToSellerResponse(seller.Result)

	return c.JSON(http.StatusOK, response)
}

// PutSellerController updates an existing seller.
func (sc *SellerController) PutSellerController(c echo.Context) error {
	var updateSellerRequest request.UpdateSellerRequest

	if err := c.Bind(&updateSellerRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to parse request body",
		})
	}

	updateSellerCommand, err := updateSellerRequest.ToUpdateSellerCommand()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid seller Id format",
		})
	}

	commandResult, err := sc.service.UpdateSeller(updateSellerCommand)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update seller",
		})
	}

	response := mapper.ToSellerResponse(commandResult.Result)

	return c.JSON(http.StatusOK, response)
}

// DeleteSellerController deletes a seller by ID.
func (sc *SellerController) DeleteSellerController(c echo.Context) error {
	// Hack: split the Id from the URL
	// For some reason c.Param("id") doesn't work here
	idRaw := c.Request().URL.Path[len("/api/v1/sellers/"):]

	id, err := uuid.Parse(idRaw)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid seller Id format",
		})
	}

	err = sc.service.DeleteSeller(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete seller",
		})
	}

	return c.NoContent(http.StatusNoContent)
}
