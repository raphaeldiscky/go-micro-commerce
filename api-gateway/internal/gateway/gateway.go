// Package gateway implements the API Gateway for the microservices architecture.
package gateway

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/middleware/metrics"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/middleware/tracing"
	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/service"
)

// Gateway represents the API Gateway.
type Gateway struct {
	logger           logger.Logger
	serviceDiscovery service.Discovery
	circuitBreaker   *service.CircuitBreakerService
	loadBalancer     service.LoadBalancer
	metrics          *metrics.Metrics
	config           *config.Config
}

// Config holds gateway configuration.
type Config struct {
	Logger           logger.Logger
	ServiceDiscovery service.Discovery
	CircuitBreaker   *service.CircuitBreakerService
	LoadBalancer     service.LoadBalancer
	Metrics          *metrics.Metrics
	Config           *config.Config
}

// NewAPIGateway creates a new API Gateway instance.
func NewAPIGateway(cfg Config) *Gateway {
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
				"service", serviceName,
				"error", err)

			return echo.NewHTTPError(http.StatusServiceUnavailable, "service unavailable")
		}

		// Execute request through circuit breaker
		result, err := gw.circuitBreaker.Execute(serviceName, func() (interface{}, error) {
			return gw.proxyRequest(c, endpoint, path)
		})

		duration := time.Since(start)

		if err != nil {
			gw.logger.Error("Circuit breaker rejected request",
				"service", serviceName,
				"error", err)

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
				"service", serviceName)

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

	// Determine final path to forward
	// If a path template is provided, replace parameters.
	// Otherwise, derive from the incoming request by removing the route prefix.
	finalPath := path
	if finalPath == "" {
		incomingPath := c.Request().URL.Path // e.g., /products/health
		routePattern := c.Path()             // e.g., /products/*
		basePrefix := strings.TrimSuffix(routePattern, "*")
		trimmedPrefix := strings.TrimRight(basePrefix, "/")
		suffix := strings.TrimPrefix(incomingPath, trimmedPrefix)

		switch suffix {
		case "", "/":
			finalPath = "/"
		default:
			if !strings.HasPrefix(suffix, "/") {
				finalPath = "/" + suffix
			} else {
				finalPath = suffix
			}
		}
	} else {
		finalPath = gw.replacePath(path, c)
	}

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

	// Add user headers
	gw.addUserHeaders(c, req)

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
			gw.logger.Warn("Failed to close response body", "error", closeErr)
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

// addUserHeaders adds user information headers to the request.
func (gw *Gateway) addUserHeaders(c echo.Context, req *http.Request) {
	// Get user information from context (set by Authorization middleware)
	if userID := c.Get(string(constant.CtxUserID)); userID != nil {
		if id, ok := userID.(uuid.UUID); ok {
			req.Header.Set(constant.XUserID, id.String())
		}
	}

	if email := c.Get(string(constant.CtxEmail)); email != nil {
		if emailStr, ok := email.(string); ok {
			req.Header.Set(constant.XEmail, emailStr)
		}
	}

	if roles := c.Get(string(constant.CtxRoles)); roles != nil {
		if rolesSlice, ok := roles.([]string); ok {
			// Join roles with comma or send as JSON
			req.Header.Set(constant.XRoles, strings.Join(rolesSlice, ","))
		}
	}

	if isActive := c.Get(string(constant.CtxIsActive)); isActive != nil {
		if active, ok := isActive.(bool); ok {
			req.Header.Set(constant.XIsActive, strconv.FormatBool(active))
		}
	}
}
