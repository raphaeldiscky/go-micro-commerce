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
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/dto"
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

// ConversationGetter is a function type for getting user conversations.
type ConversationGetter func(ctx context.Context, userID uuid.UUID, userType constant.UserType) ([]dto.ConversationResponse, error)

// ChatConnectionHandler implements the universal ConnectionHandler interface for chat.
type ChatConnectionHandler struct {
	connectionRepo     repository.ConnectionRepository
	messageRepo        repository.MessageRepository
	logger             logger.Logger
	hub                *ChatHub
	conversationGetter ConversationGetter
}

// NewChatConnectionHandler creates a new chat connection handler.
func NewChatConnectionHandler(
	connectionRepo repository.ConnectionRepository,
	messageRepo repository.MessageRepository,
	logger logger.Logger,
	hub *ChatHub,
	conversationGetter ConversationGetter,
) *ChatConnectionHandler {
	return &ChatConnectionHandler{
		connectionRepo:     connectionRepo,
		messageRepo:        messageRepo,
		logger:             logger,
		hub:                hub,
		conversationGetter: conversationGetter,
	}
}

// NewChatConnection creates a new chat-specific WebSocket connection.
func NewChatConnection(
	userID uuid.UUID,
	userType constant.UserType,
	conn *websocket.Conn,
	hub *ChatHub,
	connectionRepo repository.ConnectionRepository,
	messageRepo repository.MessageRepository,
	conversationGetter ConversationGetter,
	logger logger.Logger,
) *ChatConnection {
	handler := NewChatConnectionHandler(
		connectionRepo,
		messageRepo,
		logger,
		hub,
		conversationGetter,
	)
	config := &pkgwebsocket.ConnectionConfig{
		ReadBufferSize:  constant.WebSocketReadBufferSize,
		WriteBufferSize: constant.WebSocketWriteBufferSize,
		MaxMessageSize:  constant.WebSocketMaxMessageSize,
		PongWait:        constant.WebSocketPongWait,
		PingPeriod:      constant.WebSocketPingPeriod,
		GracePeriod:     constant.WebSocketGracePeriod,
		WriteWait:       constant.WebSocketWriteWait,
		SendBufferSize:  constant.WebSocketSendBufferSize,
	}

	baseConn := pkgwebsocket.NewBaseConnection(userID, conn, hub, handler, config, logger)

	chatConn := &ChatConnection{
		BaseConnection: baseConn,
		userType:       userType,
		connectionRepo: connectionRepo,
		logger:         logger,
	}

	// Set the actual connection instance that handlers will receive
	baseConn.SetSelf(chatConn)

	return chatConn
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

	// Auto-join user to their existing conversations
	if chatConn, ok := conn.(*ChatConnection); ok {
		if err := h.autoJoinUserConversations(chatConn); err != nil {
			h.logger.Error("Failed to auto-join conversations",
				"error", err,
				"user_id", conn.UserID())
			// Don't fail the connection for this - just log the error
		}
	}

	return nil
}

// autoJoinUserConversations automatically joins the user to all their active conversations.
func (h *ChatConnectionHandler) autoJoinUserConversations(chatConn *ChatConnection) error {
	ctx := context.Background()

	// Get user's conversations using the conversation getter function
	conversations, err := h.conversationGetter(
		ctx,
		chatConn.UserID(),
		chatConn.UserType(),
	)
	if err != nil {
		return err
	}

	// Join each conversation channel
	joinCount := 0

	for _, conv := range conversations {
		// Parse conversation ID from response
		conversationID, parseErr := uuid.Parse(conv.ID.String())
		if parseErr != nil {
			h.logger.Error("Failed to parse conversation ID",
				"error", parseErr,
				"conversation_id", conv.ID)

			continue
		}

		// Join the conversation channel in the hub
		h.hub.JoinConversation(chatConn, conversationID)

		joinCount++

		h.logger.Debug("Auto-joined conversation",
			"user_id", chatConn.UserID(),
			"conversation_id", conversationID,
			"conversation_subject", conv.Subject)
	}

	h.logger.Info("Auto-joined conversations",
		"user_id", chatConn.UserID(),
		"total_conversations", len(conversations),
		"joined_count", joinCount)

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
	// Check if this is a normal WebSocket closure
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		h.logger.Debug("Chat connection closed normally",
			"connection_id", conn.ID(),
			"user_id", conn.UserID())

		return
	}

	// Log unexpected errors
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
		"conversation_id", content.ConversationID,
		"text", content.Text,
		"message_type", content.MessageType)

	// Use the conversation ID directly from the message content
	conversationID := content.ConversationID
	ctx := context.Background()

	// Get connection as ChatConnection for user type information
	chatConn, ok := conn.(*ChatConnection)
	if !ok {
		h.logger.Error("Failed to cast connection to ChatConnection", "connection_id", conn.ID())
		return pkgwebsocket.ErrInvalidMessage
	}

	// Validate that the user is a participant in the specified conversation
	err := h.validateUserParticipation(ctx, conn.UserID(), chatConn.UserType(), conversationID)
	if err != nil {
		h.logger.Error("User not authorized for conversation",
			"user_id", conn.UserID(),
			"conversation_id", conversationID,
			"error", err)

		return pkgwebsocket.ErrInvalidMessage
	}

	// Create message entity
	messageEntity, err := entity.NewMessage(
		conversationID,
		conn.UserID(),
		content.Text,
		content.MessageType,
	)
	if err != nil {
		h.logger.Error("Failed to create message entity", "error", err)
		return err
	}

	savedMessage, err := h.messageRepo.Create(ctx, messageEntity)
	if err != nil {
		h.logger.Error("Failed to save message to database", "error", err)
		return err
	}

	h.logger.Info("Message saved to database",
		"message_id", savedMessage.ID,
		"conversation_id", conversationID,
		"sender_id", conn.UserID())

	// Broadcast message to conversation participants
	err = h.hub.BroadcastToConversation(conversationID, message, conn.UserID())
	if err != nil {
		h.logger.Error("Failed to broadcast message", "error", err)
		// Don't return error here - message is saved, broadcast failure shouldn't fail the operation
	}

	// Send delivery receipt back to sender
	receiptMsg, err := NewDeliveryReceiptMessage(
		savedMessage.ID,
		conversationID,
		conn.UserID(),
		time.Now().Unix(),
	)
	if err != nil {
		h.logger.Error("Failed to create delivery receipt", "error", err)
		// Don't return error - message was saved and broadcast successfully
	} else {
		// Send receipt directly to sender
		if err = conn.Send(receiptMsg); err != nil {
			h.logger.Error("Failed to send delivery receipt", "error", err)
		}
	}

	h.logger.Debug("Chat message handled successfully",
		"message_id", savedMessage.ID,
		"conversation_id", conversationID)

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

// validateUserParticipation validates that a user is an active participant in the specified conversation.
func (h *ChatConnectionHandler) validateUserParticipation(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
	conversationID uuid.UUID,
) error {
	// Get user's conversations to validate participation
	conversations, err := h.conversationGetter(ctx, userID, userType)
	if err != nil {
		h.logger.Error("Failed to get user conversations for validation",
			"error", err,
			"user_id", userID,
			"conversation_id", conversationID)

		return err
	}

	// Check if the user is a participant in the specified conversation
	for _, conv := range conversations {
		if conv.ID.String() == conversationID.String() {
			h.logger.Debug("User participation validated",
				"user_id", userID,
				"conversation_id", conversationID)

			return nil
		}
	}

	h.logger.Warn("User not a participant in conversation",
		"user_id", userID,
		"conversation_id", conversationID,
		"user_conversations", len(conversations))

	return pkgwebsocket.ErrInvalidMessage
}
