// Package handler provides HTTP handlers for the fulfillment service.
package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
)

// AppHandler handles application-level requests.
type AppHandler struct{}

// NewAppHandler creates a new instance of AppHandler.
func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

// Health handles health check.
func (c *AppHandler) Health(e echo.Context) error {
	return e.JSON(http.StatusOK, dto.WebResponse[any]{
		Data: map[string]any{
			"status":  "healthy",
			"service": "chat-service",
		},
		Message: "service is healthy",
	})
}

// RouteNotFound handles 404 errors.
func (c *AppHandler) RouteNotFound(e echo.Context) error {
	return e.JSON(http.StatusNotFound, dto.WebResponse[any]{
		Message: "route not found",
	})
}
