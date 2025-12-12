// Package gateway implements the API Gateway for the microservices architecture.
package gateway

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/constant"
)

const numProxyWorkers = 2

// ProxyWebSocket creates a handler that proxies WebSocket connections to a backend service.
func (gw *Gateway) ProxyWebSocket(serviceName, path string) echo.HandlerFunc {
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

		// Resolve path and build WebSocket URL
		finalPath := gw.resolvePath(c, path)

		backendURL, err := gw.buildWebSocketURL(endpoint, finalPath, c)
		if err != nil {
			gw.logger.Errorf("failed to build backend WebSocket URL: %v", err)

			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		// Extract subprotocols from client request
		clientSubprotocols := websocket.Subprotocols(c.Request())

		// Prepare response headers for WebSocket subprotocol negotiation
		responseHeader := http.Header{}
		if len(clientSubprotocols) > 0 {
			responseHeader.Set("Sec-WebSocket-Protocol", clientSubprotocols[0])
		}

		// Upgrade client connection to WebSocket
		clientConn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), responseHeader)
		if err != nil {
			gw.logger.Errorf("failed to upgrade client connection: %v", err)

			return err
		}

		defer func() {
			if closeErr := clientConn.Close(); closeErr != nil {
				gw.logger.Warn("failed to close client WebSocket connection", "error", closeErr)
			}
		}()

		// Prepare headers for backend connection
		backendHeaders := http.Header{}
		gw.prepareHeaders(c, backendHeaders, HeaderOptions{IsWebSocket: true})

		// Dial backend WebSocket
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
			return gw.handleBackendDialFailure(clientConn, serviceName, start, resp, err)
		}

		defer func() {
			if backendConn != nil {
				if closeErr := backendConn.Close(); closeErr != nil {
					gw.logger.Warn(
						"failed to close backend WebSocket connection",
						"error",
						closeErr,
					)
				}
			}
		}()

		// Record successful connection metrics
		duration := time.Since(start)
		gw.telemetry.RecordBackendRequest(
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
	gw.telemetry.RecordBackendRequest(serviceName, "WEBSOCKET", statusCode, duration)

	// Send close message to client
	closeErr := clientConn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "backend unavailable"),
	)
	if closeErr != nil {
		gw.logger.Warn("failed to send close message to client", "error", closeErr)
	}

	return echo.NewHTTPError(statusCode, "failed to connect to backend")
}

// proxyWebSocketMessages proxies messages bidirectionally between client and backend.
func (gw *Gateway) proxyWebSocketMessages(clientConn, backendConn *websocket.Conn) {
	// Set up ping handlers
	backendConn.SetPingHandler(func(appData string) error {
		gw.logger.Debug("received ping from backend, sending pong")

		err := backendConn.WriteControl(
			websocket.PongMessage,
			[]byte(appData),
			time.Now().Add(constant.WsServerWriteWait),
		)
		if err != nil {
			gw.logger.Warn("failed to send pong to backend", "error", err)
		}

		return err
	})

	clientConn.SetPingHandler(func(appData string) error {
		gw.logger.Debug("received ping from client, sending pong")

		err := clientConn.WriteControl(
			websocket.PongMessage,
			[]byte(appData),
			time.Now().Add(constant.WsServerWriteWait),
		)
		if err != nil {
			gw.logger.Warn("failed to send pong to client", "error", err)
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
					gw.logger.Warn("client WebSocket read error", "error", err)
				}

				if writeErr := backendConn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				); writeErr != nil {
					gw.logger.Warn("failed to send close to backend", "error", writeErr)
				}

				break
			}

			if err = backendConn.WriteMessage(messageType, message); err != nil {
				gw.logger.Warn("backend WebSocket write error", "error", err)

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
					gw.logger.Warn("backend WebSocket read error", "error", err)
				}

				if writeErr := clientConn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				); writeErr != nil {
					gw.logger.Warn("failed to send close to client", "error", writeErr)
				}

				break
			}

			if err = clientConn.WriteMessage(messageType, message); err != nil {
				gw.logger.Warn("client WebSocket write error", "error", err)

				break
			}
		}
	}()

	wg.Wait()
}
