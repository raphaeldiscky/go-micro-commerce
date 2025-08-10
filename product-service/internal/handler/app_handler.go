package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/dto"
)

type AppHandler struct {
}

// NewAppHandler creates a new instance of AppHandler.
func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

// Route sets up the HTTP routes for the application.
func (c *AppHandler) Route(e *echo.Echo) {
	e.HTTPErrorHandler = c.CustomHTTPErrorHandler
}

// Health handles health check.
func (h *AppHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "product-service",
	})
}

// CustomHTTPErrorHandler handles all HTTP errors including 404 and 405
func (c *AppHandler) CustomHTTPErrorHandler(err error, ctx echo.Context) {
	code := http.StatusInternalServerError
	message := "internal server error"

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		switch code {
		case http.StatusNotFound:
			message = "route not found"
		case http.StatusMethodNotAllowed:
			message = "method not allowed"
		default:
			if he.Message != nil {
				if msg, ok := he.Message.(string); ok {
					message = msg
				}
			}
		}
	}

	// Send JSON response for all errors
	if !ctx.Response().Committed {
		if err := ctx.JSON(code, dto.WebResponse[any]{
			Message: message,
		}); err != nil {
			ctx.Logger().Error(err)
		}
	}
}
