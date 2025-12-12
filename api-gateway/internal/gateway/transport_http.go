// Package gateway implements the API Gateway for the microservices architecture.
package gateway

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// ProxyToService creates a handler that proxies HTTP requests to a backend service.
func (gw *Gateway) ProxyToService(serviceName, path string) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		// Get service endpoint
		endpoint, err := gw.serviceDiscovery.GetServiceEndpoint(serviceName)
		if err != nil {
			gw.logger.Errorf("failed to get service endpoint for service %s: %v",
				serviceName, err)

			return echo.NewHTTPError(http.StatusServiceUnavailable, "service unavailable")
		}

		// Execute request through circuit breaker
		result, err := gw.circuitBreaker.Execute(serviceName, func() (any, error) {
			return gw.executeHTTPProxy(c, endpoint, path)
		})

		duration := time.Since(start)

		if err != nil {
			gw.logger.Errorf(
				"circuit breaker rejected request for service %s: %v",
				serviceName,
				err,
			)

			gw.telemetry.RecordBackendRequest(
				serviceName,
				c.Request().Method,
				http.StatusServiceUnavailable,
				duration,
			)

			return echo.NewHTTPError(http.StatusServiceUnavailable, "service circuit breaker open")
		}

		response, ok := result.(*ProxyResponse)
		if !ok {
			gw.logger.Errorf(
				"invalid response type from circuit breaker for service %s",
				serviceName,
			)

			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		// Record metrics
		gw.telemetry.RecordBackendRequest(
			serviceName,
			c.Request().Method,
			response.StatusCode,
			duration,
		)

		// Set response headers
		for key, values := range response.Headers {
			for _, value := range values {
				c.Response().Header().Add(key, value)
			}
		}

		return c.Blob(response.StatusCode, response.ContentType, response.Body)
	}
}

// executeHTTPProxy performs the actual HTTP request to the backend service.
func (gw *Gateway) executeHTTPProxy(c echo.Context, endpoint, path string) (*ProxyResponse, error) {
	// Resolve path
	finalPath := gw.resolvePath(c, path)

	// Build target URL
	targetURL, err := gw.buildTargetURL(endpoint, finalPath, c)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(
		c.Request().Context(),
		c.Request().Method,
		targetURL.String(),
		c.Request().Body,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Prepare headers
	gw.prepareHeaders(c, req.Header, HeaderOptions{})

	// Perform request
	client := &http.Client{
		Timeout: gw.config.App.TimeoutProxyRequest,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			gw.logger.Warn("failed to close response body", "error", closeErr)
		}
	}()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &ProxyResponse{
		StatusCode:  resp.StatusCode,
		Headers:     resp.Header,
		Body:        body,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
}
