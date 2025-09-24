package websocket

import (
	"context"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/pubsub"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/repository"
)

// ChatHub extends the universal BaseHub with chat-specific functionality.
type ChatHub struct {
	*pkgwebsocket.BaseHub

	logger           logger.Logger
	ConnectionRepo   repository.ConnectionRepository
	MessageRepo      repository.MessageRepository
	pubSub           *pubsub.ChatPubSub
	messagePublisher MessagePublisher
	messageParser    MessageParser
}

// NewChatHub creates a new chat-specific WebSocket hub.
func NewChatHub(
	connectionRepo repository.ConnectionRepository,
	messageRepo repository.MessageRepository,
	logger logger.Logger,
	chatPubSub *pubsub.ChatPubSub,
) *ChatHub {
	baseHub := pkgwebsocket.NewBaseHub(logger)

	var messagePublisher MessagePublisher
	if chatPubSub != nil {
		messagePublisher = NewMessagePublisher(chatPubSub)
	}

	hub := &ChatHub{
		BaseHub:          baseHub,
		ConnectionRepo:   connectionRepo,
		MessageRepo:      messageRepo,
		logger:           logger,
		pubSub:           chatPubSub,
		messagePublisher: messagePublisher,
		messageParser:    NewMessageParser(),
	}

	// Register handlers for cross-instance messages
	if chatPubSub != nil {
		hub.registerCrossInstanceHandlers()
	}

	return hub
}

// BroadcastToConversation broadcasts a message to all participants in a conversation.
func (h *ChatHub) BroadcastToConversation(
	conversationID uuid.UUID,
	message *pkgwebsocket.Message,
	excludeUserID ...uuid.UUID,
) error {
	channelName := ConversationChannel(conversationID)

	// Determine exclusion user ID
	var excludeUID *uuid.UUID
	if len(excludeUserID) > 0 {
		excludeUID = &excludeUserID[0]
	}

	// Broadcast to local connections
	var localErr error

	if excludeUID != nil {
		h.broadcastWithFilter(channelName, message, *excludeUID)
	} else {
		localErr = h.BroadcastToChannel(channelName, message)
	}

	// Publish to Redis for other instances (if pub/sub is available)
	if h.messagePublisher != nil {
		ctx := context.Background()
		if err := h.messagePublisher.PublishMessage(ctx, conversationID, message, excludeUID); err != nil {
			h.logger.Error("Failed to publish message to Redis", "error", err)
			// Don't return this error - local broadcast is more critical
		}
	}

	return localErr
}

// BroadcastToUserType broadcasts a message to all users of a specific type.
func (h *ChatHub) BroadcastToUserType(
	userType constant.UserType,
	message *pkgwebsocket.Message,
) error {
	filter := func(conn pkgwebsocket.Connection) bool {
		if chatConn, ok := conn.(*ChatConnection); ok {
			return chatConn.UserType() == userType
		}

		return false
	}

	return h.Broadcast(message, filter)
}

// JoinConversation adds a connection to a conversation channel.
func (h *ChatHub) JoinConversation(conn *ChatConnection, conversationID uuid.UUID) {
	channelName := ConversationChannel(conversationID)
	h.JoinChannel(conn, channelName)
	conn.JoinConversation(conversationID)
}

// LeaveConversation removes a connection from a conversation channel.
func (h *ChatHub) LeaveConversation(conn *ChatConnection, conversationID uuid.UUID) {
	channelName := ConversationChannel(conversationID)
	h.LeaveChannel(conn, channelName)
	conn.LeaveConversation()
}

// GetConversationConnections returns all connections in a conversation.
func (h *ChatHub) GetConversationConnections(conversationID uuid.UUID) []*ChatConnection {
	channelName := ConversationChannel(conversationID)
	universalConns := h.GetChannelConnections(channelName)

	chatConns := make([]*ChatConnection, 0, len(universalConns))
	for _, conn := range universalConns {
		if chatConn, ok := conn.(*ChatConnection); ok {
			chatConns = append(chatConns, chatConn)
		}
	}

	return chatConns
}

// GetUserTypeConnections returns all connections for a specific user type.
func (h *ChatHub) GetUserTypeConnections(userType constant.UserType) []*ChatConnection {
	connections := make([]*ChatConnection, 0)

	// Use filter function with broadcast to collect matching connections
	filter := func(conn pkgwebsocket.Connection) bool {
		if chatConn, ok := conn.(*ChatConnection); ok && chatConn.UserType() == userType {
			connections = append(connections, chatConn)
		}

		return false // Don't actually send the message
	}

	// Create a dummy message and use broadcast with filter
	dummyMessage := &pkgwebsocket.Message{}
	if err := h.BaseHub.Broadcast(dummyMessage, filter); err != nil {
		h.logger.Error("Failed to collect user type connections", "error", err)
	}

	return connections
}

// GetConnectionCount returns the total number of active connections.
func (h *ChatHub) GetConnectionCount() int {
	return h.BaseHub.GetConnectionCount()
}

// GetUserCount returns the number of unique users connected.
func (h *ChatHub) GetUserCount() int {
	return h.BaseHub.GetUserCount()
}

// GetOnlineUsers returns a list of all currently online user IDs.
func (h *ChatHub) GetOnlineUsers() []uuid.UUID {
	userIDs := make(map[uuid.UUID]bool)

	// Use filter function to collect unique user IDs
	filter := func(conn pkgwebsocket.Connection) bool {
		if conn.IsActive() {
			userIDs[conn.UserID()] = true
		}

		return false // Don't send the message
	}

	// Use broadcast with filter to iterate through connections
	if err := h.BaseHub.Broadcast(&pkgwebsocket.Message{}, filter); err != nil {
		h.logger.Error("Failed to collect online users", "error", err)
	}

	// Convert map keys to slice
	userList := make([]uuid.UUID, 0, len(userIDs))
	for userID := range userIDs {
		userList = append(userList, userID)
	}

	return userList
}

// StartRedisSubscriber starts the Redis subscriber to handle cross-instance messages.
func (h *ChatHub) StartRedisSubscriber(ctx context.Context) error {
	if h.pubSub == nil {
		h.logger.Info("Redis pub/sub not configured, skipping subscriber")
		return nil
	}

	h.logger.Info("Starting Redis subscriber for cross-instance messages")

	return h.pubSub.StartSubscriber(ctx)
}

// registerCrossInstanceHandlers registers handlers for cross-instance messages.
func (h *ChatHub) registerCrossInstanceHandlers() {
	h.pubSub.RegisterHandler(pubsub.CrossInstanceMessageTypeChat, h.handleCrossInstanceChatMessage)
	h.pubSub.RegisterHandler(
		pubsub.CrossInstanceMessageTypeTyping,
		h.handleCrossInstanceTypingMessage,
	)
	h.pubSub.RegisterHandler(
		pubsub.CrossInstanceMessageTypePresence,
		h.handleCrossInstancePresenceMessage,
	)
	h.pubSub.RegisterHandler(
		pubsub.CrossInstanceMessageTypeDeliveryReceipt,
		h.handleCrossInstanceReceiptMessage,
	)
	h.pubSub.RegisterHandler(
		pubsub.CrossInstanceMessageTypeReadReceipt,
		h.handleCrossInstanceReceiptMessage,
	)
}

// handleCrossInstanceChatMessage handles chat messages from other instances.
func (h *ChatHub) handleCrossInstanceChatMessage(
	_ context.Context,
	crossMsg *pubsub.CrossInstanceMessage,
) error {
	if crossMsg.ConversationID == nil {
		return nil
	}

	// Parse the WebSocket message from the payload
	wsMessage, err := h.messageParser.ParseWebSocketMessage(crossMsg.Payload)
	if err != nil {
		return err
	}

	// Broadcast to local connections only (no Redis re-publishing)
	return h.broadcastToLocalConversation(
		*crossMsg.ConversationID,
		wsMessage,
		crossMsg.ExcludeUserID,
	)
}

// handleCrossInstanceTypingMessage handles typing indicators from other instances.
func (h *ChatHub) handleCrossInstanceTypingMessage(
	_ context.Context,
	crossMsg *pubsub.CrossInstanceMessage,
) error {
	if crossMsg.ConversationID == nil {
		return nil
	}

	wsMessage, err := h.messageParser.ParseWebSocketMessage(crossMsg.Payload)
	if err != nil {
		return err
	}

	return h.broadcastToLocalConversation(
		*crossMsg.ConversationID,
		wsMessage,
		crossMsg.ExcludeUserID,
	)
}

// handleCrossInstancePresenceMessage handles presence updates from other instances.
func (h *ChatHub) handleCrossInstancePresenceMessage(
	_ context.Context,
	crossMsg *pubsub.CrossInstanceMessage,
) error {
	wsMessage, err := h.messageParser.ParseWebSocketMessage(crossMsg.Payload)
	if err != nil {
		return err
	}

	// Broadcast presence to all local connections
	return h.Broadcast(wsMessage, nil)
}

// handleCrossInstanceReceiptMessage handles receipt messages from other instances.
func (h *ChatHub) handleCrossInstanceReceiptMessage(
	_ context.Context,
	crossMsg *pubsub.CrossInstanceMessage,
) error {
	if crossMsg.ConversationID == nil {
		return nil
	}

	wsMessage, err := h.messageParser.ParseWebSocketMessage(crossMsg.Payload)
	if err != nil {
		return err
	}

	return h.broadcastToLocalConversation(
		*crossMsg.ConversationID,
		wsMessage,
		crossMsg.ExcludeUserID,
	)
}

// broadcastToLocalConversation broadcasts to local connections only (no Redis).
func (h *ChatHub) broadcastToLocalConversation(
	conversationID uuid.UUID,
	message *pkgwebsocket.Message,
	excludeUserID *uuid.UUID,
) error {
	channelName := ConversationChannel(conversationID)

	if excludeUserID != nil {
		filter := func(conn pkgwebsocket.Connection) bool {
			return conn.UserID() != *excludeUserID
		}

		connections := h.GetChannelConnections(channelName)
		for _, conn := range connections {
			if filter(conn) {
				if err := conn.Send(message); err != nil {
					h.logger.Error(
						"Failed to send cross-instance message to connection",
						"error",
						err,
					)
				}
			}
		}

		return nil
	}

	return h.BroadcastToChannel(channelName, message)
}

// broadcastWithFilter broadcasts a message to a channel excluding a specific user.
func (h *ChatHub) broadcastWithFilter(
	channelName string,
	message *pkgwebsocket.Message,
	excludeUserID uuid.UUID,
) {
	filter := func(conn pkgwebsocket.Connection) bool {
		return conn.UserID() != excludeUserID
	}

	connections := h.GetChannelConnections(channelName)
	for _, conn := range connections {
		if filter(conn) {
			if err := conn.Send(message); err != nil {
				h.logger.Error(
					"Failed to send message to connection",
					"error",
					err,
				)
			}
		}
	}
}

// Shutdown gracefully shuts down the ChatHub including Redis pub/sub.
func (h *ChatHub) Shutdown(ctx context.Context) error {
	h.logger.Info("Shutting down ChatHub...")

	var errs []error

	// Shutdown Redis pub/sub
	if h.pubSub != nil {
		if err := h.pubSub.Shutdown(); err != nil {
			errs = append(errs, err)
		}
	}

	// Shutdown base hub
	if err := h.BaseHub.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errs[0] // Return first error
	}

	h.logger.Info("ChatHub shutdown completed")

	return nil
}
