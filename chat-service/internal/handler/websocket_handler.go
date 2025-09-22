package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/middleware"
	chatwebsocket "github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// WebSocketHandler handles WebSocket connections for the chat service.
type WebSocketHandler struct {
	hub    *chatwebsocket.ChatHub
	logger logger.Logger
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(hub *chatwebsocket.ChatHub, logger logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
	}
}

// HandleWebSocket handles WebSocket connection upgrades.
func (h *WebSocketHandler) HandleWebSocket(c echo.Context) error {
	auth, err := middleware.AuthenticateWebSocket(c.Request())
	if err != nil {
		h.logger.Error("WebSocket authentication failed", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication failed")
	}

	if err := middleware.RequireActiveUser()(auth); err != nil {
		h.logger.Error("WebSocket authorization failed", "error", err)
		return echo.NewHTTPError(http.StatusForbidden, "authorization failed")
	}

	// Use universal websocket upgrader
	conn, err := websocket.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", "error", err)
		return err
	}

	// Create chat-specific connection
	wsConn := chatwebsocket.NewChatConnection(
		auth.UserID,
		auth.UserType,
		conn,
		h.hub,
		h.hub.ConnectionRepo, // Access through hub
		h.logger,
	)

	// Register with hub
	h.hub.Register(wsConn)

	// Start connection handling
	go wsConn.Start(c.Request().Context())

	h.logger.Info("WebSocket connection established",
		"user_id", auth.UserID,
		"user_type", auth.UserType,
		"connection_id", wsConn.ID())

	return nil
}

// HandleAdminWebSocket handles WebSocket connections specifically for admin users.
func (h *WebSocketHandler) HandleAdminWebSocket(c echo.Context) error {
	auth, err := middleware.AuthenticateWebSocket(c.Request())
	if err != nil {
		h.logger.Error("WebSocket authentication failed", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication failed")
	}

	if err := middleware.RequireActiveUser()(auth); err != nil {
		h.logger.Error("WebSocket authorization failed", "error", err)
		return echo.NewHTTPError(http.StatusForbidden, "authorization failed")
	}

	if err := middleware.RequireUserType("admin")(auth); err != nil {
		h.logger.Error("WebSocket admin authorization failed", "error", err)
		return echo.NewHTTPError(http.StatusForbidden, "admin access required")
	}

	// Use universal websocket upgrader
	conn, err := websocket.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", "error", err)
		return err
	}

	// Create chat-specific connection
	wsConn := chatwebsocket.NewChatConnection(
		auth.UserID,
		auth.UserType,
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
		"user_id", auth.UserID,
		"connection_id", wsConn.ID())

	return nil
}

// GetConnectionStats returns WebSocket connection statistics.
func (h *WebSocketHandler) GetConnectionStats(c echo.Context) error {
	stats := map[string]interface{}{
		"total_connections": h.hub.GetConnectionCount(),
		"unique_users":      h.hub.GetUserCount(),
	}

	return c.JSON(http.StatusOK, stats)
}
