// Package gateway implements the API Gateway for the microservices architecture.
package gateway

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/metrics"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/middleware/tracing"
	"github.com/raphaeldiscky/go-micro-template/api-gateway/internal/service"
)

// Gateway represents the API Gateway.
type Gateway struct {
	logger           *zap.Logger
	serviceDiscovery service.Discovery
	circuitBreaker   *service.CircuitBreakerService
	loadBalancer     service.LoadBalancer
	metrics          *metrics.Metrics
	config           *config.Config
}

// Config holds gateway configuration.
type Config struct {
	Logger           *zap.Logger
	ServiceDiscovery service.Discovery
	CircuitBreaker   *service.CircuitBreakerService
	LoadBalancer     service.LoadBalancer
	Metrics          *metrics.Metrics
	Config           *config.Config
}

// New creates a new API Gateway instance.
func New(cfg Config) *Gateway {
	return &Gateway{
		logger:           cfg.Logger,
		serviceDiscovery: cfg.ServiceDiscovery,
		circuitBreaker:   cfg.CircuitBreaker,
		loadBalancer:     cfg.LoadBalancer,
		metrics:          cfg.Metrics,
		config:           cfg.Config,
	}
}

// ProxyToService creates a handler that proxies requests to a service.
func (gw *Gateway) ProxyToService(serviceName, path string) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		// Get service endpoint
		endpoint, err := gw.serviceDiscovery.GetServiceEndpoint(serviceName)
		if err != nil {
			gw.logger.Error("Failed to get service endpoint",
				zap.String("service", serviceName),
				zap.Error(err))

			return echo.NewHTTPError(http.StatusServiceUnavailable, "service unavailable")
		}

		// Execute request through circuit breaker
		result, err := gw.circuitBreaker.Execute(serviceName, func() (interface{}, error) {
			return gw.proxyRequest(c, endpoint, path)
		})

		duration := time.Since(start)

		if err != nil {
			gw.logger.Error("Circuit breaker rejected request",
				zap.String("service", serviceName),
				zap.Error(err))

			// Record metrics
			gw.metrics.RecordGatewayRequest(
				serviceName,
				c.Request().Method,
				http.StatusServiceUnavailable,
				duration,
			)

			return echo.NewHTTPError(http.StatusServiceUnavailable, "service circuit breaker open")
		}

		response, ok := result.(*ProxyResponse)
		if !ok {
			gw.logger.Error("Invalid response type from circuit breaker",
				zap.String("service", serviceName))

			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		// Record metrics
		gw.metrics.RecordGatewayRequest(
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

// ProxyResponse represents a proxy response.
type ProxyResponse struct {
	StatusCode  int
	Headers     http.Header
	Body        []byte
	ContentType string
}

// proxyRequest performs the actual HTTP request to the service.
func (gw *Gateway) proxyRequest(c echo.Context, endpoint, path string) (*ProxyResponse, error) {
	// Build target URL
	targetURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Replace path parameters
	finalPath := gw.replacePath(path, c)
	targetURL.Path = finalPath
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

	// Copy headers (excluding hop-by-hop headers)
	gw.copyHeaders(c.Request().Header, req.Header)

	// Add tracing headers
	headers := make(map[string]string)
	tracing.InjectHeaders(c.Request().Context(), headers)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add gateway identification
	req.Header.Set("X-Gateway", "api-gateway")
	req.Header.Set("X-Forwarded-For", c.RealIP())
	req.Header.Set("X-Forwarded-Proto", c.Scheme())
	req.Header.Set("X-Forwarded-Host", c.Request().Host)

	// Perform request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			gw.logger.Warn("Failed to close response body", zap.Error(closeErr))
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

// replacePath replaces path parameters in the target path.
func (gw *Gateway) replacePath(targetPath string, c echo.Context) string {
	path := targetPath

	// Replace path parameters
	for _, param := range c.ParamNames() {
		placeholder := ":" + param
		value := c.Param(param)
		path = strings.ReplaceAll(path, placeholder, value)
	}

	return path
}

// copyHeaders copies HTTP headers, excluding hop-by-hop headers.
func (gw *Gateway) copyHeaders(src, dst http.Header) {
	hopByHopHeaders := map[string]bool{
		"Connection":          true,
		"Keep-Alive":          true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"Te":                  true,
		"Trailers":            true,
		"Transfer-Encoding":   true,
		"Upgrade":             true,
	}

	for key, values := range src {
		if !hopByHopHeaders[key] {
			for _, value := range values {
				dst.Add(key, value)
			}
		}
	}
}

// CreateReverseProxy creates a reverse proxy for a service (alternative approach).
func (gw *Gateway) CreateReverseProxy(serviceName string) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get service endpoint
		endpoint, err := gw.serviceDiscovery.GetServiceEndpoint(serviceName)
		if err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "service unavailable")
		}

		targetURL, err := url.Parse(endpoint)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadGateway, "invalid service endpoint")
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// Customize the director
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Header.Set("X-Gateway", "api-gateway")
			req.Header.Set("X-Forwarded-For", c.RealIP())
		}

		// Handle errors
		proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
			gw.logger.Error("Proxy error",
				zap.String("service", serviceName),
				zap.Error(err))
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		}

		proxy.ServeHTTP(c.Response(), c.Request())

		return nil
	}
}
