// Package gateway implements the API Gateway for the microservices architecture.
package gateway

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
)

// ProxyToConnectRPC creates a handler that proxies Connect-RPC requests to a service,
// preserving the full service method path.
func (gw *Gateway) ProxyToConnectRPC(serviceName string) echo.HandlerFunc {
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
			return gw.executeConnectRPCProxy(c, endpoint)
		})

		duration := time.Since(start)

		if err != nil {
			gw.logger.Errorf("circuit breaker rejected request for service %s: %v",
				serviceName, err)

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
			gw.logger.Error("invalid response type from circuit breaker",
				"service", serviceName)

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

// executeConnectRPCProxy performs the actual HTTP request to the Connect-RPC service,
// preserving the full method path.
func (gw *Gateway) executeConnectRPCProxy(c echo.Context, endpoint string) (*ProxyResponse, error) {
	// Build target URL - preserve full path for Connect-RPC
	targetURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// For Connect-RPC, preserve the full path from the original request
	targetURL.Path = c.Request().URL.Path
	targetURL.RawQuery = c.Request().URL.RawQuery

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
		return nil, fmt.Errorf("failed to perform request: %w", err)
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
