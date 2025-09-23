// Package websocket provides chat-specific WebSocket implementations.
package websocket

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

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
	hub            *ChatHub
}

// NewChatConnectionHandler creates a new chat connection handler.
func NewChatConnectionHandler(
	connectionRepo repository.ConnectionRepository,
	logger logger.Logger,
	hub *ChatHub,
) *ChatConnectionHandler {
	return &ChatConnectionHandler{
		connectionRepo: connectionRepo,
		logger:         logger,
		hub:            hub,
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
	handler := NewChatConnectionHandler(connectionRepo, logger, hub)
	config := &pkgwebsocket.ConnectionConfig{
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
	// Create database connection record using background context
	// since the WebSocket connection is long-lived but the HTTP request context is short-lived
	dbCtx := context.Background()

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

	if _, err := c.connectionRepo.Create(dbCtx, connEntity); err != nil {
		c.logger.Error("Failed to create connection record", "error", err)
	}

	c.BaseConnection.Start(ctx)
}

// OnConnect handles connection establishment.
func (h *ChatConnectionHandler) OnConnect(conn pkgwebsocket.Connection) error {
	h.logger.Info("Chat connection established",
		"connection_id", conn.ID(),
		"user_id", conn.UserID())

	// No need to send welcome message - WebSocket upgrade response itself
	// signals successful connection to the client
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
	case ChatMessageTypeChat:
		// Handle chat messages
		return h.handleChatMessage(conn, message)
	case ChatMessageTypeTyping:
		// Handle typing indicators
		return h.handleTypingMessage(conn, message)
	case ChatMessageTypePresence:
		// Handle presence updates
		return h.handlePresenceMessage(conn, message)
	case ChatMessageTypeDeliveryReceipt:
		// Handle delivery receipts
		return h.handleDeliveryReceiptMessage(conn, message)
	case ChatMessageTypeReadReceipt:
		// Handle read receipts
		return h.handleReadReceiptMessage(conn, message)
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
	// Parse chat content
	var content ChatContent
	if err := message.ParseContent(&content); err != nil {
		h.logger.Error("Failed to parse chat message", "error", err)
		return err
	}

	h.logger.Debug("Chat message received",
		"connection_id", conn.ID(),
		"message_id", message.ID,
		"text", content.Text,
		"message_type", content.MessageType)

	// Here you would typically:
	// 1. Validate the message
	// 2. Store it in the database
	// 3. Broadcast to conversation participants
	// 4. Send delivery receipts

	return nil
}

// handleTypingMessage handles typing indicator messages.
func (h *ChatConnectionHandler) handleTypingMessage(
	conn pkgwebsocket.Connection,
	message *pkgwebsocket.Message,
) error {
	var content TypingContent
	if err := message.ParseContent(&content); err != nil {
		h.logger.Error("Failed to parse typing message", "error", err)
		return err
	}

	h.logger.Debug("Typing indicator received",
		"connection_id", conn.ID(),
		"message_id", message.ID,
		"is_typing", content.IsTyping)

	// Broadcast typing indicator to conversation participants (excluding the sender)
	if chatConn, ok := conn.(*ChatConnection); ok && chatConn.ConversationID() != nil {
		return h.hub.BroadcastToConversation(*chatConn.ConversationID(), message, conn.UserID())
	}

	return nil
}

// handlePresenceMessage handles presence update messages.
func (h *ChatConnectionHandler) handlePresenceMessage(
	conn pkgwebsocket.Connection,
	message *pkgwebsocket.Message,
) error {
	var content PresenceContent
	if err := message.ParseContent(&content); err != nil {
		h.logger.Error("Failed to parse presence message", "error", err)
		return err
	}

	h.logger.Debug("Presence update received",
		"connection_id", conn.ID(),
		"message_id", message.ID,
		"user_id", content.UserID,
		"status", content.Status,
		"event", content.Event)

	// Broadcast presence update to all relevant connections
	// This could be to conversation participants or globally depending on requirements
	err := h.hub.Broadcast(message, nil)
	if err != nil {
		return err
	}

	return nil
}

// handleDeliveryReceiptMessage handles delivery receipt messages.
func (h *ChatConnectionHandler) handleDeliveryReceiptMessage(
	conn pkgwebsocket.Connection,
	message *pkgwebsocket.Message,
) error {
	var content DeliveryReceiptContent
	if err := message.ParseContent(&content); err != nil {
		h.logger.Error("Failed to parse delivery receipt message", "error", err)
		return err
	}

	h.logger.Debug("Delivery receipt received",
		"connection_id", conn.ID(),
		"message_id", message.ID,
		"original_message_id", content.MessageID,
		"conversation_id", content.ConversationID,
		"recipient_id", content.RecipientID)

	// Send delivery receipt back to the original sender
	return h.hub.BroadcastToConversation(content.ConversationID, message, conn.UserID())
}

// handleReadReceiptMessage handles read receipt messages.
func (h *ChatConnectionHandler) handleReadReceiptMessage(
	conn pkgwebsocket.Connection,
	message *pkgwebsocket.Message,
) error {
	var content ReadReceiptContent
	if err := message.ParseContent(&content); err != nil {
		h.logger.Error("Failed to parse read receipt message", "error", err)
		return err
	}

	h.logger.Debug("Read receipt received",
		"connection_id", conn.ID(),
		"message_id", message.ID,
		"original_message_id", content.MessageID,
		"conversation_id", content.ConversationID,
		"reader_id", content.ReaderID)

	// Send read receipt back to the original sender
	return h.hub.BroadcastToConversation(content.ConversationID, message, conn.UserID())
}
