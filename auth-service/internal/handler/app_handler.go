package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-template/pkg/dto"
)

// AppHandler handles application-level requests.
type AppHandler struct{}

// NewAppHandler creates a new instance of AppHandler.
func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

// Health handles health check.
func (c *AppHandler) Health(e echo.Context) error {
	return e.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "auth-service",
	})
}

// CustomHTTPErrorHandler handles all HTTP errors including 404 and 405.
func (c *AppHandler) CustomHTTPErrorHandler(err error, ctx echo.Context) {
	code := http.StatusInternalServerError
	message := "internal server error"

	he := &echo.HTTPError{}
	if errors.As(err, &he) {
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
