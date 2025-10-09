// Package gateway implements the API Gateway for the microservices architecture.
package gateway

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
	metrics          *metrics.Metrics
	config           *config.Config
}

// Config holds gateway configuration.
type Config struct {
	Logger           logger.Logger
	ServiceDiscovery service.Discovery
	CircuitBreaker   *service.CircuitBreakerService
	Metrics          *metrics.Metrics
	Config           *config.Config
}

// NewAPIGateway creates a new API Gateway instance.
func NewAPIGateway(cfg Config) *Gateway {
	return &Gateway{
		logger:           cfg.Logger,
		serviceDiscovery: cfg.ServiceDiscovery,
		circuitBreaker:   cfg.CircuitBreaker,
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
			gw.logger.Errorf("failed to get service endpoint for service %s: %v",
				serviceName, err)

			return echo.NewHTTPError(http.StatusServiceUnavailable, "service unavailable")
		}

		// Execute request through circuit breaker
		result, err := gw.circuitBreaker.Execute(serviceName, func() (any, error) {
			return gw.proxyRequest(c, endpoint, path)
		})

		duration := time.Since(start)

		if err != nil {
			gw.logger.Errorf(
				"Circuit breaker rejected request for service %s: %v",
				serviceName,
				err,
			)

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
			gw.logger.Errorf(
				"invalid response type from circuit breaker for service %s",
				serviceName,
			)

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
		Timeout: gw.config.App.TimeoutProxyRequest,
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
// For WebSocket connections, preserves Connection and Upgrade headers.
func (gw *Gateway) copyHeaders(src, dst http.Header) {
	gw.copyHeadersWithOptions(src, dst, false)
}

// copyHeadersWithOptions copies HTTP headers with optional WebSocket support.
func (gw *Gateway) copyHeadersWithOptions(src, dst http.Header, isWebSocket bool) {
	hopByHopHeaders := map[string]bool{
		"Keep-Alive":          true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"Te":                  true,
		"Trailers":            true,
		"Transfer-Encoding":   true,
		"Connection":          true, // Always hop-by-hop, never copy
		"Upgrade":             true, // Always hop-by-hop, never copy
	}

	if isWebSocket {
		// For WebSocket connections, additionally exclude handshake headers
		// The websocket.Dialer generates these headers automatically
		hopByHopHeaders["Sec-Websocket-Key"] = true
		hopByHopHeaders["Sec-Websocket-Version"] = true
		hopByHopHeaders["Sec-Websocket-Extensions"] = true
		hopByHopHeaders["Sec-Websocket-Protocol"] = true
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
	if userID := c.Get(string(constant.CtxKeyUserID)); userID != nil {
		if id, ok := userID.(uuid.UUID); ok {
			req.Header.Set(constant.XUserID, id.String())
		}
	}

	if email := c.Get(string(constant.CtxKeyEmail)); email != nil {
		if emailStr, ok := email.(string); ok {
			req.Header.Set(constant.XEmail, emailStr)
		}
	}

	if roles := c.Get(string(constant.CtxKeyRoles)); roles != nil {
		if rolesSlice, ok := roles.([]string); ok {
			// Join roles with comma or send as JSON
			req.Header.Set(constant.XRoles, strings.Join(rolesSlice, ","))
		}
	}

	// Add client metadata for audit logging and security
	req.Header.Set(constant.XClientIP, c.RealIP())
	req.Header.Set(constant.XUserAgent, c.Request().UserAgent())
}

// ProxyToConnectRPC creates a handler that proxies Connect-RPC requests to a service,
// preserving the full service method path.
func (gw *Gateway) ProxyToConnectRPC(serviceName string) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		// Get service endpoint
		endpoint, err := gw.serviceDiscovery.GetServiceEndpoint(serviceName)
		if err != nil {
			gw.logger.Errorf("failed to get service endpoint for service %s: %v",
				serviceName,
				err)

			return echo.NewHTTPError(http.StatusServiceUnavailable, "service unavailable")
		}

		// Execute request through circuit breaker, preserving full path
		result, err := gw.circuitBreaker.Execute(serviceName, func() (any, error) {
			return gw.proxyConnectRPCRequest(c, endpoint)
		})

		duration := time.Since(start)

		if err != nil {
			gw.logger.Errorf("circuit breaker rejected request for service %s: %v",
				serviceName, err)

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

// proxyConnectRPCRequest performs the actual HTTP request to the Connect-RPC service,
// preserving the full method path.
func (gw *Gateway) proxyConnectRPCRequest(c echo.Context, endpoint string) (*ProxyResponse, error) {
	// Build target URL
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
		Timeout: gw.config.App.TimeoutProxyRequest,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
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

// ProxyWebSocket creates a handler that proxies WebSocket connections to a service.
func (gw *Gateway) ProxyWebSocket(serviceName, path string) echo.HandlerFunc {
	// WebSocket upgrader with permissive settings for proxying
	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true // Allow all origins for proxying
		},
		ReadBufferSize:  constant.WsServerReadBufferSize,
		WriteBufferSize: constant.WsServerWriteBufferSize,
	}

	return func(c echo.Context) error {
		start := time.Now()

		// Get service endpoint
		endpoint, err := gw.serviceDiscovery.GetServiceEndpoint(serviceName)
		if err != nil {
			gw.logger.Errorf("failed to get service endpoint for service %s: %v",
				serviceName, err)

			return echo.NewHTTPError(http.StatusServiceUnavailable, "service unavailable")
		}

		// Build backend WebSocket URL
		backendURL, err := gw.buildBackendWebSocketURL(endpoint, path, c)
		if err != nil {
			gw.logger.Errorf("failed to build backend WebSocket URL: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		// Extract subprotocols from client request (e.g., graphql-transport-ws)
		// This is needed for GraphQL subscriptions protocol negotiation
		clientSubprotocols := websocket.Subprotocols(c.Request())

		// Prepare response headers for WebSocket subprotocol negotiation
		// For WebSocket proxying, we need to accept the client's requested subprotocol
		// and forward it to the backend. The backend will validate if it's supported.
		responseHeader := http.Header{}
		if len(clientSubprotocols) > 0 {
			// Accept the first requested subprotocol (typically graphql-transport-ws)
			// This header is required for graphql-ws clients to accept the connection
			responseHeader.Set("Sec-WebSocket-Protocol", clientSubprotocols[0])
		}

		// Upgrade client connection to WebSocket
		// Use c.Response().Writer to get the underlying http.ResponseWriter that supports hijacking
		clientConn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), responseHeader)
		if err != nil {
			gw.logger.Errorf("failed to upgrade client connection: %v", err)

			return err
		}

		defer func() {
			if closeErr := clientConn.Close(); closeErr != nil {
				gw.logger.Warn("Failed to close client WebSocket connection", "error", closeErr)
			}
		}()

		// Prepare headers for backend connection
		backendHeaders := gw.prepareBackendHeaders(c)

		// Dial backend WebSocket with subprotocols
		// The Subprotocols field ensures proper protocol negotiation (e.g., graphql-transport-ws)
		dialer := websocket.Dialer{
			HandshakeTimeout: gw.config.App.TimeoutProxyRequest,
			Subprotocols:     clientSubprotocols,
		}

		backendConn, resp, err := dialer.Dial(backendURL, backendHeaders)

		// Close response body if present
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				gw.logger.Errorf("failed to close backend response body: %v", closeErr)
			}
		}

		// Handle connection failures
		if err != nil || backendConn == nil {
			return gw.handleBackendDialFailure(
				clientConn,
				serviceName,
				start,
				resp,
				err,
			)
		}

		defer func() {
			if backendConn != nil {
				if closeErr := backendConn.Close(); closeErr != nil {
					gw.logger.Warn(
						"Failed to close backend WebSocket connection",
						"error",
						closeErr,
					)
				}
			}
		}()

		// Record successful connection metrics
		duration := time.Since(start)
		gw.metrics.RecordGatewayRequest(
			serviceName,
			"WEBSOCKET",
			http.StatusSwitchingProtocols,
			duration,
		)

		// Proxy messages bidirectionally
		gw.proxyWebSocketMessages(clientConn, backendConn)

		return nil
	}
}

// buildBackendWebSocketURL builds the backend WebSocket URL from endpoint and path.
func (gw *Gateway) buildBackendWebSocketURL(endpoint, path string, c echo.Context) (string, error) {
	// Parse endpoint URL
	targetURL, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Convert http(s) to ws(s)
	scheme := "ws"
	if targetURL.Scheme == "https" {
		scheme = "wss"
	}

	// Determine final path
	finalPath := path
	if finalPath == "" {
		incomingPath := c.Request().URL.Path
		routePattern := c.Path()
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

	// Build WebSocket URL
	wsURL := fmt.Sprintf("%s://%s%s", scheme, targetURL.Host, finalPath)
	if c.Request().URL.RawQuery != "" {
		wsURL += "?" + c.Request().URL.RawQuery
	}

	return wsURL, nil
}

// prepareBackendHeaders prepares headers for backend WebSocket connection.
func (gw *Gateway) prepareBackendHeaders(c echo.Context) http.Header {
	backendHeaders := http.Header{}
	gw.copyHeadersWithOptions(c.Request().Header, backendHeaders, true)

	// Add user headers from context
	if userID := c.Get(string(constant.CtxKeyUserID)); userID != nil {
		if id, ok := userID.(uuid.UUID); ok {
			backendHeaders.Set(constant.XUserID, id.String())
		}
	}

	if email := c.Get(string(constant.CtxKeyEmail)); email != nil {
		if emailStr, ok := email.(string); ok {
			backendHeaders.Set(constant.XEmail, emailStr)
		}
	}

	if roles := c.Get(string(constant.CtxKeyRoles)); roles != nil {
		if rolesSlice, ok := roles.([]string); ok {
			backendHeaders.Set(constant.XRoles, strings.Join(rolesSlice, ","))
		}
	}

	// Add tracing headers
	tracingHeaders := make(map[string]string)
	tracing.InjectHeaders(c.Request().Context(), tracingHeaders)

	for key, value := range tracingHeaders {
		backendHeaders.Set(key, value)
	}

	// Add gateway identification and client metadata
	backendHeaders.Set("X-Gateway", "api-gateway")
	backendHeaders.Set("X-Forwarded-For", c.RealIP())
	backendHeaders.Set("X-Forwarded-Proto", c.Scheme())
	backendHeaders.Set("X-Forwarded-Host", c.Request().Host)
	backendHeaders.Set(constant.XClientIP, c.RealIP())
	backendHeaders.Set(constant.XUserAgent, c.Request().UserAgent())

	return backendHeaders
}

const numProxyWorkers = 2

// proxyWebSocketMessages proxies messages bidirectionally between client and backend.
func (gw *Gateway) proxyWebSocketMessages(clientConn, backendConn *websocket.Conn) {
	// Set up ping handlers to automatically respond with pong (keep connections alive)
	// The proxy transparently handles ping/pong without managing deadlines
	backendConn.SetPingHandler(func(appData string) error {
		gw.logger.Debug("Received ping from backend, sending pong")

		err := backendConn.WriteControl(
			websocket.PongMessage,
			[]byte(appData),
			time.Now().Add(constant.WsServerWriteWait),
		)
		if err != nil {
			gw.logger.Warn("Failed to send pong to backend", "error", err)
		}

		return err
	})

	clientConn.SetPingHandler(func(appData string) error {
		gw.logger.Debug("Received ping from client, sending pong")

		err := clientConn.WriteControl(
			websocket.PongMessage,
			[]byte(appData),
			time.Now().Add(constant.WsServerWriteWait),
		)
		if err != nil {
			gw.logger.Warn("Failed to send pong to client", "error", err)
		}

		return err
	})

	var wg sync.WaitGroup

	wg.Add(numProxyWorkers)

	// Client to backend
	go func() {
		defer wg.Done()

		for {
			messageType, message, err := clientConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,
					websocket.CloseNormalClosure,
				) {
					gw.logger.Warn("Client WebSocket read error", "error", err)
				}

				// Close backend connection
				err = backendConn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				if err != nil {
					gw.logger.Warn("Backend WebSocket write error", "error", err)
				}

				break
			}

			if err = backendConn.WriteMessage(messageType, message); err != nil {
				gw.logger.Warn("Backend WebSocket write error", "error", err)

				break
			}
		}
	}()

	// Backend to client
	go func() {
		defer wg.Done()

		for {
			messageType, message, err := backendConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,
					websocket.CloseNormalClosure,
				) {
					gw.logger.Warn("Backend WebSocket read error", "error", err)
				}

				// Close client connection
				err = clientConn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				if err != nil {
					gw.logger.Warn("Client WebSocket write error", "error", err)
				}

				break
			}

			if err = clientConn.WriteMessage(messageType, message); err != nil {
				gw.logger.Warn("Client WebSocket write error", "error", err)

				break
			}
		}
	}()

	// Wait for both goroutines to complete
	wg.Wait()
}

// handleBackendDialFailure handles failures when dialing the backend WebSocket.
func (gw *Gateway) handleBackendDialFailure(
	clientConn *websocket.Conn,
	serviceName string,
	start time.Time,
	resp *http.Response,
	err error,
) error {
	duration := time.Since(start)
	statusCode := http.StatusBadGateway

	if resp != nil {
		statusCode = resp.StatusCode
	}

	if err != nil {
		gw.logger.Errorf("failed to dial backend WebSocket: %v", err)
	} else {
		gw.logger.Error("backend WebSocket connection is nil despite no error")
	}

	// Record metrics
	gw.metrics.RecordGatewayRequest(serviceName, "WEBSOCKET", statusCode, duration)

	// Send close message to client
	closeErr := clientConn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(
			websocket.CloseInternalServerErr,
			"backend unavailable",
		),
	)
	if closeErr != nil {
		gw.logger.Warn("Failed to send close message to client", "error", closeErr)
	}

	return echo.NewHTTPError(statusCode, "failed to connect to backend")
}
