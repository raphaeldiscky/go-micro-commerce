package websocket

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/rediseventbus"

	redispkg "github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/event"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/repository"
)

// ChatHub extends the universal BaseHub with chat-specific functionality.
type ChatHub struct {
	*pkgwebsocket.BaseHub

	logger         logger.Logger
	ConnectionRepo repository.ConnectionRepository
	MessageRepo    repository.MessageRepository
	eventBus       rediseventbus.EventBus
	eventHandler   *event.ChatEventHandler
	instanceID     string
	activeChannels map[string]int // channel -> connection count
	channelMutex   sync.RWMutex
}

// NewChatHub creates a new chat-specific WebSocket hub.
func NewChatHub(
	connectionRepo repository.ConnectionRepository,
	messageRepo repository.MessageRepository,
	logger logger.Logger,
	instanceID string,
) *ChatHub {
	baseHub := pkgwebsocket.NewBaseHub(logger)

	hub := &ChatHub{
		BaseHub:        baseHub,
		ConnectionRepo: connectionRepo,
		MessageRepo:    messageRepo,
		logger:         logger,
		instanceID:     instanceID,
		activeChannels: make(map[string]int),
	}

	return hub
}

// SetEventBus sets the event bus and initializes event handlers.
func (h *ChatHub) SetEventBus(bus rediseventbus.EventBus) {
	h.eventBus = bus

	if bus == nil {
		h.logger.Info("Event bus not configured, cross-instance messaging disabled")
		return
	}

	// Create event handler
	h.eventHandler = event.NewChatEventHandler(h.logger)

	// Register event handlers
	h.eventHandler.SetChatMessageHandler(h.handleChatMessageEvent)
	h.eventHandler.SetTypingIndicatorHandler(h.handleTypingIndicatorEvent)
	h.eventHandler.SetPresenceUpdateHandler(h.handlePresenceUpdateEvent)
	h.eventHandler.SetDeliveryReceiptHandler(h.handleDeliveryReceiptEvent)
	h.eventHandler.SetReadReceiptHandler(h.handleReadReceiptEvent)

	h.logger.Info("Event bus configured successfully", "instance_id", h.instanceID)
}

// BroadcastToConversation broadcasts a message to all participants in a conversation.
func (h *ChatHub) BroadcastToConversation(
	conversationID uuid.UUID,
	message *pkgwebsocket.Message,
	excludeUserID ...uuid.UUID,
) error {
	channelName := redispkg.ConversationChannel(conversationID)

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

	// Publish to EventBus for other instances (if available)
	if h.eventBus != nil {
		ctx := context.Background()
		chatEvent := &event.ChatMessageEvent{
			ConversationID: conversationID,
			Message:        message,
			ExcludeUserID:  excludeUID,
		}

		if err := h.publishEvent(ctx, channelName, event.TypeChatMessage, chatEvent); err != nil {
			h.logger.Error("Failed to publish event to Redis", "error", err)
			// Don't return this error - local broadcast is more critical
		}
	}

	return localErr
}

// BroadcastTypingIndicator broadcasts a typing indicator to a conversation.
func (h *ChatHub) BroadcastTypingIndicator(
	conversationID uuid.UUID,
	message *pkgwebsocket.Message,
	excludeUserID uuid.UUID,
) error {
	channelName := redispkg.ConversationChannel(conversationID)

	// Broadcast locally
	h.broadcastWithFilter(channelName, message, excludeUserID)

	// Publish to EventBus
	if h.eventBus != nil {
		ctx := context.Background()
		typingEvent := &event.TypingIndicatorEvent{
			ConversationID: conversationID,
			Message:        message,
			ExcludeUserID:  &excludeUserID,
		}

		if err := h.publishEvent(ctx, channelName, event.TypeTypingIndicator, typingEvent); err != nil {
			h.logger.Error("Failed to publish typing indicator", "error", err)
		}
	}

	return nil
}

// BroadcastPresenceUpdate broadcasts a presence update to all connections.
func (h *ChatHub) BroadcastPresenceUpdate(
	userID uuid.UUID,
	message *pkgwebsocket.Message,
) error {
	// Broadcast locally
	if err := h.Broadcast(message, nil); err != nil {
		return err
	}

	// Publish to EventBus
	if h.eventBus != nil {
		ctx := context.Background()
		presenceEvent := &event.PresenceUpdateEvent{
			UserID:  userID,
			Message: message,
		}

		channelName := redispkg.UserChannel(userID)
		if err := h.publishEvent(ctx, channelName, event.TypePresenceUpdate, presenceEvent); err != nil {
			h.logger.Error("Failed to publish presence update", "error", err)
		}
	}

	return nil
}

// BroadcastDeliveryReceipt broadcasts a delivery receipt to a conversation.
func (h *ChatHub) BroadcastDeliveryReceipt(
	conversationID uuid.UUID,
	message *pkgwebsocket.Message,
	excludeUserID uuid.UUID,
) error {
	channelName := redispkg.ConversationChannel(conversationID)

	// Broadcast locally
	h.broadcastWithFilter(channelName, message, excludeUserID)

	// Publish to EventBus
	if h.eventBus != nil {
		ctx := context.Background()
		receiptEvent := &event.DeliveryReceiptEvent{
			ConversationID: conversationID,
			Message:        message,
			ExcludeUserID:  &excludeUserID,
		}

		if err := h.publishEvent(ctx, channelName, event.TypeDeliveryReceipt, receiptEvent); err != nil {
			h.logger.Error("Failed to publish delivery receipt", "error", err)
		}
	}

	return nil
}

// BroadcastReadReceipt broadcasts a read receipt to a conversation.
func (h *ChatHub) BroadcastReadReceipt(
	conversationID uuid.UUID,
	message *pkgwebsocket.Message,
	excludeUserID uuid.UUID,
) error {
	channelName := redispkg.ConversationChannel(conversationID)

	// Broadcast locally
	h.broadcastWithFilter(channelName, message, excludeUserID)

	// Publish to EventBus
	if h.eventBus != nil {
		ctx := context.Background()
		receiptEvent := &event.ReadReceiptEvent{
			ConversationID: conversationID,
			Message:        message,
			ExcludeUserID:  &excludeUserID,
		}

		if err := h.publishEvent(ctx, channelName, event.TypeReadReceipt, receiptEvent); err != nil {
			h.logger.Error("Failed to publish read receipt", "error", err)
		}
	}

	return nil
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

// subscribeToChannelIfNeeded subscribes to a Redis channel if this is the first connection.
func (h *ChatHub) subscribeToChannelIfNeeded(channelName string, conversationID uuid.UUID) {
	if h.eventBus == nil {
		return
	}

	h.channelMutex.Lock()
	defer h.channelMutex.Unlock()

	// Subscribe to Redis if first connection for this conversation
	if h.activeChannels[channelName] == 0 {
		err := h.eventBus.Subscribe(channelName, h.eventHandler.HandleEvent)
		if err != nil {
			h.logger.Error("Failed to subscribe to channel",
				"channel", channelName,
				"conversation_id", conversationID,
				"error", err)
		} else {
			h.logger.Info("Subscribed to conversation channel",
				"channel", channelName,
				"conversation_id", conversationID,
				"instance_id", h.instanceID)
		}
	}

	h.activeChannels[channelName]++
}

// JoinConversation adds a connection to a conversation channel with dynamic Redis subscription.
func (h *ChatHub) JoinConversation(conn *ChatConnection, conversationID uuid.UUID) {
	channelName := redispkg.ConversationChannel(conversationID)

	// Handle Redis subscription with dynamic subscription
	h.subscribeToChannelIfNeeded(channelName, conversationID)

	// Join local hub channel
	h.JoinChannel(conn, channelName)
	conn.JoinConversation(conversationID)
}

// unsubscribeFromChannelIfNeeded unsubscribes from a Redis channel if this is the last connection.
func (h *ChatHub) unsubscribeFromChannelIfNeeded(channelName string, conversationID uuid.UUID) {
	if h.eventBus == nil {
		return
	}

	h.channelMutex.Lock()
	defer h.channelMutex.Unlock()

	h.activeChannels[channelName]--

	// Unsubscribe from Redis if no more connections
	if h.activeChannels[channelName] <= 0 {
		err := h.eventBus.Unsubscribe(channelName)
		if err != nil {
			h.logger.Error("Failed to unsubscribe from channel",
				"channel", channelName,
				"conversation_id", conversationID,
				"error", err)
		} else {
			h.logger.Info("Unsubscribed from conversation channel",
				"channel", channelName,
				"conversation_id", conversationID,
				"instance_id", h.instanceID)
		}

		delete(h.activeChannels, channelName)
	}
}

// LeaveConversation removes a connection from a conversation channel with dynamic Redis unsubscription.
func (h *ChatHub) LeaveConversation(conn *ChatConnection, conversationID uuid.UUID) {
	channelName := redispkg.ConversationChannel(conversationID)

	// Handle Redis unsubscription
	h.unsubscribeFromChannelIfNeeded(channelName, conversationID)

	// Leave local hub channel
	h.LeaveChannel(conn, channelName)
	conn.LeaveConversation()
}

// GetConversationConnections returns all connections in a conversation.
func (h *ChatHub) GetConversationConnections(conversationID uuid.UUID) []*ChatConnection {
	channelName := redispkg.ConversationChannel(conversationID)
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

// GetActiveChannelCount returns the number of active Redis subscriptions.
func (h *ChatHub) GetActiveChannelCount() int {
	h.channelMutex.RLock()
	defer h.channelMutex.RUnlock()

	return len(h.activeChannels)
}

// publishEvent publishes an event to the event bus.
func (h *ChatHub) publishEvent(
	ctx context.Context,
	channel string,
	eventType string,
	payload any,
) error {
	baseEvent, err := rediseventbus.NewBaseEvent(h.instanceID, eventType, payload)
	if err != nil {
		return err
	}

	return h.eventBus.Publish(ctx, channel, baseEvent)
}

// Event handlers for cross-instance events.

// handleChatMessageEvent handles chat message events from other instances.
func (h *ChatHub) handleChatMessageEvent(
	_ context.Context,
	e *event.ChatMessageEvent,
) error {
	h.logger.Debug("Handling chat message event from another instance",
		"conversation_id", e.ConversationID)

	return h.broadcastToLocalConversation(
		e.ConversationID,
		e.Message,
		e.ExcludeUserID,
	)
}

// handleTypingIndicatorEvent handles typing indicator events from other instances.
func (h *ChatHub) handleTypingIndicatorEvent(
	_ context.Context,
	e *event.TypingIndicatorEvent,
) error {
	h.logger.Debug("Handling typing indicator event from another instance",
		"conversation_id", e.ConversationID)

	return h.broadcastToLocalConversation(
		e.ConversationID,
		e.Message,
		e.ExcludeUserID,
	)
}

// handlePresenceUpdateEvent handles presence update events from other instances.
func (h *ChatHub) handlePresenceUpdateEvent(
	_ context.Context,
	e *event.PresenceUpdateEvent,
) error {
	h.logger.Debug("Handling presence update event from another instance",
		"user_id", e.UserID)

	// Broadcast presence to all local connections
	return h.Broadcast(e.Message, nil)
}

// handleDeliveryReceiptEvent handles delivery receipt events from other instances.
func (h *ChatHub) handleDeliveryReceiptEvent(
	_ context.Context,
	e *event.DeliveryReceiptEvent,
) error {
	h.logger.Debug("Handling delivery receipt event from another instance",
		"conversation_id", e.ConversationID)

	return h.broadcastToLocalConversation(
		e.ConversationID,
		e.Message,
		e.ExcludeUserID,
	)
}

// handleReadReceiptEvent handles read receipt events from other instances.
func (h *ChatHub) handleReadReceiptEvent(
	_ context.Context,
	e *event.ReadReceiptEvent,
) error {
	h.logger.Debug("Handling read receipt event from another instance",
		"conversation_id", e.ConversationID)

	return h.broadcastToLocalConversation(
		e.ConversationID,
		e.Message,
		e.ExcludeUserID,
	)
}

// broadcastToLocalConversation broadcasts to local connections only (no Redis).
func (h *ChatHub) broadcastToLocalConversation(
	conversationID uuid.UUID,
	message *pkgwebsocket.Message,
	excludeUserID *uuid.UUID,
) error {
	channelName := redispkg.ConversationChannel(conversationID)

	if excludeUserID != nil {
		h.broadcastWithFilter(channelName, message, *excludeUserID)
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

// Shutdown gracefully shuts down the ChatHub including event bus.
func (h *ChatHub) Shutdown(ctx context.Context) error {
	h.logger.Info("Shutting down ChatHub...",
		"active_subscriptions", h.GetActiveChannelCount(),
		"active_connections", h.GetConnectionCount())

	var errs []error

	// Shutdown event bus
	if h.eventBus != nil {
		if err := h.eventBus.Close(); err != nil {
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
