package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	chatwebsocket "github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// WebSocketHandler handles WebSocket connections for the chat service.
type WebSocketHandler struct {
	hub    *chatwebsocket.ChatHub
	logger logger.Logger
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(hub *chatwebsocket.ChatHub, logger logger.Logger) *WebSocketHandler {
	if hub == nil {
		if logger != nil {
			logger.Fatal("WebSocket hub cannot be nil")
		} else {
			panic("WebSocket hub cannot be nil")
		}
	}

	if logger == nil {
		panic("Logger cannot be nil")
	}

	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
	}
}

// HandleWebSocket handles WebSocket connection upgrades for regular users.
func (h *WebSocketHandler) HandleWebSocket(c echo.Context) error {
	// Check if handler is properly initialized
	if h == nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"WebSocket handler not initialized",
		)
	}

	// Check if hub is nil to prevent panic
	if h.hub == nil {
		h.logger.Error("WebSocket hub is nil")
		return echo.NewHTTPError(http.StatusInternalServerError, "WebSocket service not available")
	}

	// Check if context and request are valid
	if c == nil || c.Request() == nil {
		h.logger.Error("Invalid request context")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	// Use universal websocket upgrader
	conn, err := websocket.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", "error", err)
		return err
	}

	// Check if hub connection repository is accessible
	if h.hub.ConnectionRepo == nil {
		h.logger.Error("WebSocket hub connection repository is nil")

		if closeErr := conn.Close(); closeErr != nil {
			h.logger.Error("Failed to close WebSocket connection", "error", closeErr)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "WebSocket service misconfigured")
	}

	roles := echoutils.GetRolesFromContext(c)
	userID := echoutils.GetUserIDFromContext(c)

	// Create chat-specific connection
	wsConn := chatwebsocket.NewChatConnection(
		userID,
		constant.UserType(roles[0]),
		conn,
		h.hub,
		h.hub.ConnectionRepo,
		h.logger,
	)

	// Register with hub
	h.hub.Register(wsConn)

	// Start connection handling
	go wsConn.Start(c.Request().Context())

	h.logger.Info("WebSocket connection established",
		"user_id", userID,
		"connection_id", wsConn.ID())

	return nil
}

// HandleAdminWebSocket handles WebSocket connections specifically for admin users.
func (h *WebSocketHandler) HandleAdminWebSocket(c echo.Context) error {
	// Check if handler is properly initialized
	if h == nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"WebSocket handler not initialized",
		)
	}

	// Check if hub is nil to prevent panic
	if h.hub == nil {
		h.logger.Error("WebSocket hub is nil")
		return echo.NewHTTPError(http.StatusInternalServerError, "WebSocket service not available")
	}

	// Check if context and request are valid
	if c == nil || c.Request() == nil {
		h.logger.Error("Invalid request context")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	// Use universal websocket upgrader
	conn, err := websocket.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", "error", err)
		return err
	}

	// Check if hub connection repository is accessible
	if h.hub.ConnectionRepo == nil {
		h.logger.Error("WebSocket hub connection repository is nil")

		if closeErr := conn.Close(); closeErr != nil {
			h.logger.Error("Failed to close WebSocket connection", "error", closeErr)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "WebSocket service misconfigured")
	}

	roles := echoutils.GetRolesFromContext(c)
	userID := echoutils.GetUserIDFromContext(c)

	// Create chat-specific connection
	wsConn := chatwebsocket.NewChatConnection(
		userID,
		constant.UserType(roles[0]),
		conn,
		h.hub,
		h.hub.ConnectionRepo,
		h.logger,
	)

	// Register with hub
	h.hub.Register(wsConn)

	// Start connection handling
	go wsConn.Start(c.Request().Context())

	h.logger.Info("Admin WebSocket connection established",
		"user_id", userID,
		"connection_id", wsConn.ID())

	return nil
}

// GetConnectionStats returns WebSocket connection statistics.
func (h *WebSocketHandler) GetConnectionStats(c echo.Context) error {
	// Check if handler is properly initialized
	if h == nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"WebSocket handler not initialized",
		)
	}

	// Check if hub is nil to prevent panic
	if h.hub == nil {
		h.logger.Error("WebSocket hub is nil")
		return echo.NewHTTPError(http.StatusInternalServerError, "WebSocket service not available")
	}

	// Check if context is valid
	if c == nil {
		h.logger.Error("Invalid request context")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	// Add defensive checks for hub methods
	var totalConnections, uniqueUsers int

	// Safely get connection count
	func() {
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("Panic in GetConnectionCount", "error", r)

				totalConnections = 0
			}
		}()

		totalConnections = h.hub.GetConnectionCount()
	}()

	// Safely get user count
	func() {
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("Panic in GetUserCount", "error", r)

				uniqueUsers = 0
			}
		}()

		uniqueUsers = h.hub.GetUserCount()
	}()

	stats := map[string]any{
		"total_connections": totalConnections,
		"unique_users":      uniqueUsers,
	}

	return c.JSON(http.StatusOK, stats)
}

// WebSocketHealth handles websocket health check.
func (h *WebSocketHandler) WebSocketHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, dto.WebResponse[any]{
		Data: map[string]any{
			"status":  "healthy",
			"service": "chat-websocket",
		},
		Message: "websocket is healthy",
	})
}
