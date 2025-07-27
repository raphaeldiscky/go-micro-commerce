package handlers

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/application/dto"
	"github.com/raphaeldiscky/go-ddd-template/services/seller-service/internal/application/services"
)

// SellerHandler handles HTTP requests for seller operations
type SellerHandler struct {
	sellerService services.SellerServiceInterface
}

// NewSellerHandler creates a new instance of SellerHandler
func NewSellerHandler(sellerService services.SellerServiceInterface) *SellerHandler {
	return &SellerHandler{
		sellerService: sellerService,
	}
}

// CreateSeller handles POST /sellers
func (h *SellerHandler) CreateSeller(c echo.Context) error {
	var req dto.CreateSellerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate request
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}
	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}
	if req.Phone == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Phone is required"})
	}
	if req.Address == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Address is required"})
	}

	seller, err := h.sellerService.CreateSeller(c.Request().Context(), req)
	if err != nil {
		if err.Error() == "seller with email "+req.Email+" already exists" {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, seller)
}

// GetSeller handles GET /sellers/:id
func (h *SellerHandler) GetSeller(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid seller ID"})
	}

	seller, err := h.sellerService.GetSeller(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "seller not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Seller not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, seller)
}

// GetSellerByEmail handles GET /sellers/email/:email
func (h *SellerHandler) GetSellerByEmail(c echo.Context) error {
	email := c.Param("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}

	seller, err := h.sellerService.GetSellerByEmail(c.Request().Context(), email)
	if err != nil {
		if err.Error() == "seller not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Seller not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, seller)
}

// GetSellers handles GET /sellers
func (h *SellerHandler) GetSellers(c echo.Context) error {
	var req dto.GetSellersRequest

	// Parse query parameters
	limitParam := c.QueryParam("limit")
	if limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil && limit > 0 {
			req.Limit = limit
		}
	}
	if req.Limit == 0 {
		req.Limit = 10 // Default limit
	}

	offsetParam := c.QueryParam("offset")
	if offsetParam != "" {
		if offset, err := strconv.Atoi(offsetParam); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	activeOnlyParam := c.QueryParam("active_only")
	if activeOnlyParam == "true" {
		req.ActiveOnly = true
	}

	sellers, err := h.sellerService.GetSellers(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, sellers)
}

// UpdateSeller handles PUT /sellers/:id
func (h *SellerHandler) UpdateSeller(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid seller ID"})
	}

	var reqBody struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate request
	if reqBody.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}
	if reqBody.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}
	if reqBody.Phone == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Phone is required"})
	}
	if reqBody.Address == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Address is required"})
	}

	req := dto.UpdateSellerRequest{
		Id:      id,
		Name:    reqBody.Name,
		Email:   reqBody.Email,
		Phone:   reqBody.Phone,
		Address: reqBody.Address,
	}

	seller, err := h.sellerService.UpdateSeller(c.Request().Context(), req)
	if err != nil {
		if err.Error() == "seller not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Seller not found"})
		}
		if err.Error() == "seller with email "+reqBody.Email+" already exists" {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, seller)
}

// UpdateSellerStatus handles PATCH /sellers/:id/status
func (h *SellerHandler) UpdateSellerStatus(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid seller ID"})
	}

	var reqBody struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	req := dto.SellerStatusRequest{
		Id:       id,
		IsActive: reqBody.IsActive,
	}

	seller, err := h.sellerService.UpdateSellerStatus(c.Request().Context(), req)
	if err != nil {
		if err.Error() == "seller not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Seller not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, seller)
}

// DeleteSeller handles DELETE /sellers/:id
func (h *SellerHandler) DeleteSeller(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid seller ID"})
	}

	err = h.sellerService.DeleteSeller(c.Request().Context(), id)
	if err != nil {
		if err.Error() == "seller not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Seller not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// RegisterRoutes registers all seller routes
func (h *SellerHandler) RegisterRoutes(e *echo.Echo) {
	sellerGroup := e.Group("/api/v1/sellers")

	sellerGroup.POST("", h.CreateSeller)
	sellerGroup.GET("", h.GetSellers)
	sellerGroup.GET("/:id", h.GetSeller)
	sellerGroup.GET("/email/:email", h.GetSellerByEmail)
	sellerGroup.PUT("/:id", h.UpdateSeller)
	sellerGroup.PATCH("/:id/status", h.UpdateSellerStatus)
	sellerGroup.DELETE("/:id", h.DeleteSeller)
}
