// Package websocket provides chat-specific WebSocket implementations.
package websocket

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgconfig "github.com/raphaeldiscky/go-micro-commerce/pkg/config"
	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/repository"
)

// ChatConnection extends the universal BaseConnection with chat-specific functionality.
type ChatConnection struct {
	*pkgwebsocket.BaseConnection

	userType       constant.UserType
	conversationID *uuid.UUID
	connectionRepo repository.ConnectionRepository
	logger         logger.Logger
}

// ChatConnectionHandler implements the universal ConnectionHandler interface for chat.
type ChatConnectionHandler struct {
	connectionRepo repository.ConnectionRepository
	logger         logger.Logger
}

// NewChatConnectionHandler creates a new chat connection handler.
func NewChatConnectionHandler(
	connectionRepo repository.ConnectionRepository,
	logger logger.Logger,
) *ChatConnectionHandler {
	return &ChatConnectionHandler{
		connectionRepo: connectionRepo,
		logger:         logger,
	}
}

// NewChatConnection creates a new chat-specific WebSocket connection.
func NewChatConnection(
	userID uuid.UUID,
	userType constant.UserType,
	conn *websocket.Conn,
	hub *ChatHub,
	connectionRepo repository.ConnectionRepository,
	logger logger.Logger,
) *ChatConnection {
	handler := NewChatConnectionHandler(connectionRepo, logger)
	config := &pkgconfig.WebsocketServerConfig{
		ReadBufferSize:  constant.WebSocketReadBufferSize,
		WriteBufferSize: constant.WebSocketWriteBufferSize,
		MaxMessageSize:  constant.WebSocketMaxMessageSize,
		PongWait:        constant.WebSocketPongWait * time.Second,
		PingPeriod:      constant.WebSocketPingPeriod * time.Second,
		WriteWait:       constant.WebSocketWriteWait * time.Second,
		SendBufferSize:  constant.WebSocketSendBufferSize,
	}

	baseConn := pkgwebsocket.NewBaseConnection(userID, conn, hub, handler, config, logger)

	return &ChatConnection{
		BaseConnection: baseConn,
		userType:       userType,
		connectionRepo: connectionRepo,
		logger:         logger,
	}
}

// UserType returns the user type for this connection.
func (c *ChatConnection) UserType() constant.UserType {
	return c.userType
}

// ConversationID returns the current conversation ID (if any).
func (c *ChatConnection) ConversationID() *uuid.UUID {
	return c.conversationID
}

// JoinConversation joins a conversation.
func (c *ChatConnection) JoinConversation(conversationID uuid.UUID) {
	c.conversationID = &conversationID
}

// LeaveConversation leaves the current conversation.
func (c *ChatConnection) LeaveConversation() {
	c.conversationID = nil
}

// Start starts the chat connection with database persistence.
func (c *ChatConnection) Start(ctx context.Context) {
	// Create database connection record
	connEntity := &entity.Connection{
		ID:            c.ID(),
		UserID:        c.UserID(),
		ConnectionID:  c.ID().String(),
		SocketID:      c.ID().String(),
		UserAgent:     nil, // Could be extracted from request headers
		IPAddress:     nil, // Could be extracted from request
		ConnectedAt:   time.Now(),
		LastHeartbeat: time.Now(),
		IsActive:      true,
	}

	if _, err := c.connectionRepo.Create(ctx, connEntity); err != nil {
		c.logger.Error("Failed to create connection record", "error", err)
	}

	c.BaseConnection.Start(ctx)
}

// OnConnect handles connection establishment.
func (h *ChatConnectionHandler) OnConnect(conn pkgwebsocket.Connection) error {
	h.logger.Info("Chat connection established",
		"connection_id", conn.ID(),
		"user_id", conn.UserID())

	return nil
}

// OnDisconnect handles connection closure.
func (h *ChatConnectionHandler) OnDisconnect(conn pkgwebsocket.Connection) {
	// Mark connection as inactive in database
	if err := h.connectionRepo.MarkAsInactive(context.Background(), conn.ID().String()); err != nil {
		h.logger.Error("Failed to mark connection as inactive", "error", err)
	}

	h.logger.Info("Chat connection closed",
		"connection_id", conn.ID(),
		"user_id", conn.UserID())
}

// OnMessage handles incoming messages.
func (h *ChatConnectionHandler) OnMessage(
	conn pkgwebsocket.Connection,
	message *pkgwebsocket.Message,
) error {
	// Update heartbeat in database
	if err := h.connectionRepo.UpdateHeartbeat(context.Background(), conn.ID().String()); err != nil {
		h.logger.Warn("Failed to update heartbeat", "error", err)
	}

	h.logger.Debug("Message received",
		"connection_id", conn.ID(),
		"message_type", message.Type,
		"message_id", message.ID)

	// Handle chat-specific message types here
	switch message.Type {
	case pkgwebsocket.MessageTypeHeartbeat:
		// Heartbeat messages are handled automatically
		return nil
	case pkgwebsocket.MessageTypeCustom:
		// Handle chat-specific messages
		return h.handleChatMessage(conn, message)
	case pkgwebsocket.MessageTypeError:
		// Handle error messages
		h.logger.Error("Received error message", "type", message.Type)
		return nil
	case pkgwebsocket.MessageTypeSystem:
		// Handle system messages
		h.logger.Info("Received system message", "type", message.Type)
		return nil
	default:
		h.logger.Warn("Unknown message type", "type", message.Type)
		return pkgwebsocket.ErrInvalidMessage
	}
}

// OnError handles connection errors.
func (h *ChatConnectionHandler) OnError(conn pkgwebsocket.Connection, err error) {
	h.logger.Error("Chat connection error",
		"connection_id", conn.ID(),
		"user_id", conn.UserID(),
		"error", err)
}

// handleChatMessage handles chat-specific messages.
func (h *ChatConnectionHandler) handleChatMessage(
	conn pkgwebsocket.Connection,
	message *pkgwebsocket.Message,
) error {
	// This would handle chat messages, typing indicators, etc.
	// For now, just log the message
	h.logger.Debug("Chat message received",
		"connection_id", conn.ID(),
		"message_id", message.ID)

	return nil
}
