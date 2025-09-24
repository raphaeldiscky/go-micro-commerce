package handler

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/service"
	chatwebsocket "github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// WebSocketHandler handles WebSocket connections for the chat service.
type WebSocketHandler struct {
	hub               *chatwebsocket.ChatHub
	logger            logger.Logger
	config            *config.WebSocketServerConfig
	connectionService service.ConnectionService
	chatService       service.ChatService
}

// NewWebSocketHandler creates a new WebSocket handler.
func NewWebSocketHandler(
	hub *chatwebsocket.ChatHub,
	logger logger.Logger,
	config *config.WebSocketServerConfig,
	connectionService service.ConnectionService,
	chatService service.ChatService,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:               hub,
		logger:            logger,
		config:            config,
		connectionService: connectionService,
		chatService:       chatService,
	}
}

// createChatConnection creates a new chat connection from the Echo context.
func (h *WebSocketHandler) createChatConnection(
	c echo.Context,
) (*chatwebsocket.ChatConnection, error) {
	conn, err := pkgwebsocket.Upgrade(c.Response(), c.Request(), pkgwebsocket.UpgraderConfig{
		ReadBufferSize:  h.config.ReadBufferSize,
		WriteBufferSize: h.config.WriteBufferSize,
		CheckOrigin: func(_ *http.Request) bool {
			return true // Allow all origins for development
		},
		Subprotocols: nil,
	})
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", "error", err)
		return nil, err
	}

	// Require ticket-based authentication for all WebSocket connections
	ticket := c.QueryParam("ticket")
	if ticket == "" {
		h.logger.Error("Missing ticket parameter for WebSocket connection")

		if closeErr := conn.Close(); closeErr != nil {
			h.logger.Error("Failed to close connection", "error", closeErr)
		}

		return nil, echo.NewHTTPError(http.StatusUnauthorized, "ticket parameter is required")
	}

	return h.createConnectionFromTicket(ticket, conn)
}

// createConnectionFromTicket creates a connection using a connection ticket.
func (h *WebSocketHandler) createConnectionFromTicket(
	ticket string,
	conn *websocket.Conn,
) (*chatwebsocket.ChatConnection, error) {
	claims, err := h.connectionService.ValidateConnectionTicket(context.Background(), ticket)
	if err != nil {
		h.logger.Error("Failed to validate connection ticket", "error", err)

		if closeErr := conn.Close(); closeErr != nil {
			h.logger.Error("Failed to close connection", "error", closeErr)
		}

		return nil, err
	}

	h.logger.Info("Creating connection from ticket",
		"user_id", claims.UserID,
		"user_type", claims.UserType)

	return chatwebsocket.NewChatConnection(
		claims.UserID,
		claims.UserType,
		conn,
		h.hub,
		h.hub.ConnectionRepo,
		h.hub.MessageRepo,
		h.chatService.GetUserConversations,
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

	h.logger.Info("WebSocket connection established",
		"user_id", wsConn.UserID(),
		"user_type", wsConn.UserType(),
		"connection_id", wsConn.ID())

	// Start connection handling - this blocks until connection closes
	// The WebSocket upgrade response has already been sent, so the client
	// will receive confirmation through the standard WebSocket handshake
	wsConn.Start(context.Background())

	h.logger.Info("WebSocket connection closed",
		"user_id", wsConn.UserID(),
		"connection_id", wsConn.ID())

	return nil
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
