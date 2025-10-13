package handler

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
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

	// Require auth token for all WebSocket connections
	authToken := c.QueryParam("token")
	if authToken == "" {
		h.logger.Error("Missing token parameter for WebSocket connection")

		if closeErr := conn.Close(); closeErr != nil {
			h.logger.Error("Failed to close connection", "error", closeErr)
		}

		return nil, echo.NewHTTPError(http.StatusUnauthorized, "token parameter is required")
	}

	return h.createConnectionFromAuthToken(authToken, conn)
}

// createConnectionFromAuthToken creates a connection using an auth service JWT.
func (h *WebSocketHandler) createConnectionFromAuthToken(
	token string,
	conn *websocket.Conn,
) (*chatwebsocket.ChatConnection, error) {
	claims, err := h.connectionService.ValidateAuthToken(context.Background(), token)
	if err != nil {
		h.logger.Error("Failed to validate auth token", "error", err)

		if closeErr := conn.Close(); closeErr != nil {
			h.logger.Error("Failed to close connection", "error", closeErr)
		}

		return nil, err
	}

	// Determine user type from roles
	userType := h.determineUserTypeFromRoles(claims.Roles)

	h.logger.Info("Creating connection from auth token",
		"user_id", claims.UserID,
		"email", claims.Email,
		"user_type", userType)

	userID := uuid.MustParse(claims.UserID)

	return chatwebsocket.NewChatConnection(
		userID,
		userType,
		conn,
		h.hub,
		h.hub.ConnectionRepo,
		h.hub.MessageRepo,
		h.chatService.GetUserConversations,
		h.logger,
	), nil
}

// determineUserTypeFromRoles determines user type from JWT roles.
func (h *WebSocketHandler) determineUserTypeFromRoles(roles []string) constant.UserType {
	// Log the roles for debugging
	h.logger.Debug("Determining user type from roles", "roles", roles)

	// Prioritize admin role - check against both string and pkg constant
	for _, role := range roles {
		if role == string(constant.UserTypeAdmin) {
			h.logger.Debug("User identified as admin", "matched_role", role)
			return constant.UserTypeAdmin
		}
	}

	// Check for user role
	for _, role := range roles {
		if role == string(constant.UserTypeUser) {
			h.logger.Debug("User identified as regular user", "matched_role", role)
			return constant.UserTypeUser
		}
	}

	// If we reach here, the roles contain unexpected values
	// Default to 'user' for safety, but log a warning
	h.logger.Warn("Unexpected user roles in JWT, defaulting to user type",
		"roles", roles,
		"default_type", constant.UserTypeUser)

	return constant.UserTypeUser
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

// Health handles websocket health check.
func (h *WebSocketHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, pkgdto.WebResponse[any, any]{
		Data: map[string]any{
			"status":  "healthy",
			"service": "chat-service-ws",
		},
		Message: "service is healthy",
	})
}
