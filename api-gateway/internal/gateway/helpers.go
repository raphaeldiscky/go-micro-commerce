// Package gateway implements the API Gateway for the microservices architecture.
package gateway

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

// HeaderOptions configures header preparation behavior.
type HeaderOptions struct {
	IsWebSocket  bool
	SkipHopByHop bool
}

// isHopByHopHeader returns true if the header should not be forwarded to backends.
func isHopByHopHeader(header string) bool {
	switch header {
	case "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
		"Te", "Trailers", "Transfer-Encoding", "Connection", "Upgrade":
		return true
	default:
		return false
	}
}

// isWebSocketExcludeHeader returns true if the header should be excluded for WebSocket.
func isWebSocketExcludeHeader(header string) bool {
	switch header {
	case "Sec-Websocket-Key", "Sec-Websocket-Version",
		"Sec-Websocket-Extensions", "Sec-Websocket-Protocol":
		return true
	default:
		return false
	}
}

// resolvePath determines the final path to forward to the backend service.
// If templatePath is provided, it replaces path parameters.
// Otherwise, it derives the path from the incoming request by removing the route prefix.
func (gw *Gateway) resolvePath(c echo.Context, templatePath string) string {
	if templatePath != "" {
		return gw.replacePath(templatePath, c)
	}

	// Derive path from incoming request
	incomingPath := c.Request().URL.Path
	routePattern := c.Path()
	basePrefix := strings.TrimSuffix(routePattern, "*")
	trimmedPrefix := strings.TrimRight(basePrefix, "/")
	suffix := strings.TrimPrefix(incomingPath, trimmedPrefix)

	switch suffix {
	case "", "/":
		return "/"
	default:
		if !strings.HasPrefix(suffix, "/") {
			return "/" + suffix
		}

		return suffix
	}
}

// replacePath replaces path parameters in the target path with actual values.
func (gw *Gateway) replacePath(targetPath string, c echo.Context) string {
	path := targetPath

	for _, param := range c.ParamNames() {
		placeholder := ":" + param
		value := c.Param(param)
		path = strings.ReplaceAll(path, placeholder, value)
	}

	return path
}

// buildTargetURL builds the target URL from endpoint and resolved path.
func (gw *Gateway) buildTargetURL(endpoint, resolvedPath string, c echo.Context) (*url.URL, error) {
	targetURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	targetURL.Path = resolvedPath
	targetURL.RawQuery = c.Request().URL.RawQuery

	return targetURL, nil
}

// buildWebSocketURL builds a WebSocket URL from an HTTP endpoint.
func (gw *Gateway) buildWebSocketURL(
	endpoint, resolvedPath string,
	c echo.Context,
) (string, error) {
	targetURL, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	// Convert http(s) to ws(s)
	scheme := "ws"
	if targetURL.Scheme == "https" {
		scheme = "wss"
	}

	wsURL := scheme + "://" + targetURL.Host + resolvedPath
	if c.Request().URL.RawQuery != "" {
		wsURL += "?" + c.Request().URL.RawQuery
	}

	return wsURL, nil
}

// prepareHeaders prepares all headers for the proxy request.
// This includes copying source headers, adding user context, tracing, and gateway identification.
func (gw *Gateway) prepareHeaders(c echo.Context, dst http.Header, opts HeaderOptions) {
	// Copy headers from source, excluding hop-by-hop headers
	gw.copyHeaders(c.Request().Header, dst, opts)

	// Add user context headers
	gw.addUserHeaders(c, dst)

	// Add tracing headers
	gw.addTracingHeaders(c, dst)

	// Add gateway identification
	gw.addGatewayHeaders(c, dst)
}

// copyHeaders copies HTTP headers, excluding hop-by-hop headers.
func (gw *Gateway) copyHeaders(src, dst http.Header, opts HeaderOptions) {
	for key, values := range src {
		// Skip hop-by-hop headers
		if isHopByHopHeader(key) {
			continue
		}

		// Skip WebSocket-specific headers if applicable
		if opts.IsWebSocket && isWebSocketExcludeHeader(key) {
			continue
		}

		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// addUserHeaders adds user context headers from the Echo context.
func (gw *Gateway) addUserHeaders(c echo.Context, dst http.Header) {
	if userID := c.Get(string(constant.CtxKeyUserID)); userID != nil {
		if id, ok := userID.(uuid.UUID); ok {
			dst.Set(constant.XUserID, id.String())
		}
	}

	if email := c.Get(string(constant.CtxKeyEmail)); email != nil {
		if emailStr, ok := email.(string); ok {
			dst.Set(constant.XEmail, emailStr)
		}
	}

	if roles := c.Get(string(constant.CtxKeyRoles)); roles != nil {
		if rolesSlice, ok := roles.([]string); ok {
			dst.Set(constant.XRoles, strings.Join(rolesSlice, ","))
		}
	}

	// Add client metadata
	dst.Set(constant.XClientIP, c.RealIP())
	dst.Set(constant.XUserAgent, c.Request().UserAgent())
}

// addTracingHeaders adds distributed tracing headers.
func (gw *Gateway) addTracingHeaders(c echo.Context, dst http.Header) {
	if gw.telemetry == nil {
		return
	}

	headers := make(map[string]string)
	gw.telemetry.InjectHeaders(c.Request().Context(), headers)

	for key, value := range headers {
		dst.Set(key, value)
	}
}

// addGatewayHeaders adds gateway identification headers.
func (gw *Gateway) addGatewayHeaders(c echo.Context, dst http.Header) {
	dst.Set("X-Gateway", "api-gateway")
	dst.Set("X-Forwarded-For", c.RealIP())
	dst.Set("X-Forwarded-Proto", c.Scheme())
	dst.Set("X-Forwarded-Host", c.Request().Host)
}

// shouldSkipHeaderForSSE returns true for headers that should not be copied for SSE streaming.
func shouldSkipHeaderForSSE(key string) bool {
	skipHeaders := []string{
		"Transfer-Encoding",
		"Content-Length",
		"Connection",
		"Access-Control-",
	}

	for _, skip := range skipHeaders {
		if strings.EqualFold(key, skip) || strings.HasPrefix(key, skip) {
			return true
		}
	}

	return false
}
