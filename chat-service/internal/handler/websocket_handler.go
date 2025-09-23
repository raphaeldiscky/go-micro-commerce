package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/dto"
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

// createChatConnection creates a new chat connection from the Echo context.
func (h *WebSocketHandler) createChatConnection(
	c echo.Context,
) (*chatwebsocket.ChatConnection, error) {
	conn, err := websocket.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", "error", err)
		return nil, err
	}

	roles := echoutils.GetRolesFromContext(c)
	userID := echoutils.GetUserIDFromContext(c)

	return chatwebsocket.NewChatConnection(
		userID,
		constant.UserType(roles[0]),
		conn,
		h.hub,
		h.hub.ConnectionRepo,
		h.logger,
	), nil
}

// HandleWebSocket handles WebSocket connection upgrades for all users.
func (h *WebSocketHandler) HandleWebSocket(c echo.Context) error {
	wsConn, err := h.createChatConnection(c)
	if err != nil {
		return err
	}

	// Register with hub and start connection handling
	h.hub.Register(wsConn)

	go wsConn.Start(c.Request().Context())

	h.logger.Info("WebSocket connection established",
		"user_id", wsConn.UserID(),
		"user_type", wsConn.UserType(),
		"connection_id", wsConn.ID())

	return nil
}

// HandleAdminWebSocket handles WebSocket connections specifically for admin users.
// This is kept for backward compatibility but uses the same logic as HandleWebSocket.
func (h *WebSocketHandler) HandleAdminWebSocket(c echo.Context) error {
	return h.HandleWebSocket(c)
}

// GetConnectionStats returns WebSocket connection statistics.
func (h *WebSocketHandler) GetConnectionStats(c echo.Context) error {
	stats := dto.ConnectionStatsResponse{
		TotalConnections: h.hub.GetConnectionCount(),
		UniqueUsers:      h.hub.GetUserCount(),
	}

	return echoutils.ResponseOK(c, stats)
}

// WebSocketHealth handles websocket health check.
func (h *WebSocketHandler) WebSocketHealth(c echo.Context) error {
	healthStatus := dto.HealthStatusResponse{
		Status:  "healthy",
		Service: "chat-websocket",
	}

	return echoutils.ResponseOK(c, healthStatus)
}
