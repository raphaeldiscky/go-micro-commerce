package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/raphaeldiscky/go-micro-template/pkg/httperror"
	"github.com/raphaeldiscky/go-micro-template/pkg/logger"
)

// Logger middleware logs incoming requests and their duration.
func Logger(lgr logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			path := c.Request().URL.Path

			// Call the next handler in the chain.
			err := next(c)

			params := map[string]any{
				"status_code": c.Response().Status,
				"client_ip":   c.RealIP(),
				"method":      c.Request().Method,
				"latency":     time.Since(start).String(),
				"path":        path,
			}

			// If there's no error, log the request as successful.
			if err == nil {
				lgr.WithFields(params).Info("incoming request")

				return nil
			}

			// If an error occurred, handle it and log the details.
			handleErrorAndLog(lgr, params, err, c)

			return err
		}
	}
}

// handleErrorAndLog processes the error, sets the correct status code,
// and logs the error details.
func handleErrorAndLog(lgr logger.Logger, params map[string]any, err error, c echo.Context) {
	var (
		validationErrors validator.ValidationErrors
		httpResponseErr  *httperror.ResponseError
		echoHTTPError    *echo.HTTPError
	)

	// Use `errors.As` to check for specific error types and update the status code accordingly.
	switch {
	case errors.As(err, &validationErrors):
		// This handles validation errors. The status code is http.StatusBadRequest (400).
		c.Response().Status = http.StatusBadRequest
		params["status_code"] = http.StatusBadRequest
		params["error"] = validationErrors.Error()
	case errors.As(err, &httpResponseErr):
		// This handles custom HTTP response errors.
		c.Response().Status = httpResponseErr.GetCode()
		params["status_code"] = httpResponseErr.GetCode()
		params["error"] = httpResponseErr.OriginalError().Error()
	case errors.As(err, &echoHTTPError):
		// This handles standard Echo HTTP errors.
		c.Response().Status = echoHTTPError.Code
		params["status_code"] = echoHTTPError.Code
		params["error"] = echoHTTPError.Error()
	default:
		// For all other errors, default to a 500 Internal Server Error.
		c.Response().Status = http.StatusInternalServerError
		params["status_code"] = http.StatusInternalServerError
		params["error"] = err.Error()
	}

	lgr.WithFields(params).Error("got error")
}
