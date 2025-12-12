// Package gateway implements the API Gateway for the microservices architecture.
package gateway

import (
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// ProxySSE creates a handler that proxies Server-Sent Events to a backend service.
// SSE requires streaming without buffering and no timeouts.
func (gw *Gateway) ProxySSE(serviceName, path string) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		// Get service endpoint
		endpoint, err := gw.serviceDiscovery.GetServiceEndpoint(serviceName)
		if err != nil {
			gw.logger.Errorf("failed to get service endpoint for service %s: %v",
				serviceName, err)

			return echo.NewHTTPError(http.StatusServiceUnavailable, "service unavailable")
		}

		// Resolve path and build target URL
		finalPath := gw.resolvePath(c, path)

		targetURL, err := gw.buildTargetURL(endpoint, finalPath, c)
		if err != nil {
			gw.logger.Errorf("invalid endpoint URL: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		// Create request with no timeout (SSE is long-lived)
		req, err := http.NewRequestWithContext(
			c.Request().Context(),
			c.Request().Method,
			targetURL.String(),
			c.Request().Body,
		)
		if err != nil {
			gw.logger.Errorf("failed to create request: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		// Prepare headers
		gw.prepareHeaders(c, req.Header, HeaderOptions{})

		// Create HTTP client with no timeout for SSE streaming
		client := &http.Client{
			Timeout: 0, // No timeout for long-lived SSE connections
		}

		// Perform request
		resp, err := client.Do(req)
		if err != nil {
			gw.logger.Errorf("failed to connect to backend for SSE: %v", err)

			duration := time.Since(start)
			gw.telemetry.RecordBackendRequest(
				serviceName,
				c.Request().Method,
				http.StatusBadGateway,
				duration,
			)

			return echo.NewHTTPError(http.StatusBadGateway, "failed to connect to backend")
		}

		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				gw.logger.Warn("failed to close SSE response body", "error", closeErr)
			}
		}()

		// Record initial connection metrics
		duration := time.Since(start)
		gw.telemetry.RecordBackendRequest(
			serviceName,
			c.Request().Method,
			resp.StatusCode,
			duration,
		)

		// Copy response headers from backend (except problematic ones)
		for key, values := range resp.Header {
			if shouldSkipHeaderForSSE(key) {
				continue
			}

			for _, value := range values {
				c.Response().Header().Add(key, value)
			}
		}

		// Disable buffering for SSE streaming
		c.Response().Header().Set("X-Accel-Buffering", "no")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")

		// Write status code
		c.Response().WriteHeader(resp.StatusCode)

		// Flush headers immediately
		flusher, ok := c.Response().Writer.(http.Flusher)
		if ok && flusher != nil {
			flusher.Flush()
		}

		// Stream response body with automatic flushing
		fw := &flushWriter{
			w:       c.Response().Writer,
			flusher: flusher,
		}

		_, err = io.Copy(fw, resp.Body)
		if err != nil {
			// Connection closed by client or backend is normal for SSE
			gw.logger.Debug("SSE connection closed", "error", err)

			return nil
		}

		return nil
	}
}

// flushWriter wraps http.ResponseWriter and flushes after each write.
type flushWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// Write writes data and immediately flushes it to the client.
func (fw *flushWriter) Write(p []byte) (int, error) {
	n, err := fw.w.Write(p)
	if err != nil {
		return n, err
	}

	if fw.flusher != nil {
		fw.flusher.Flush()
	}

	return n, nil
}
