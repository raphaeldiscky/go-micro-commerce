package subscription

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/pubsub"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// Manager manages GraphQL subscriptions by bridging WebSocket events to GraphQL channels.
type Manager struct {
	logger         logger.Logger
	subscriptions  map[string]*conversationSubscription
	userSubs       map[uuid.UUID]*userSubscription
	mu             sync.RWMutex
	Hub            *websocket.ChatHub
	pubSub         *pubsub.ChatPubSub
	eventConverter *EventConverter
}

// conversationSubscription represents a subscription to conversation events.
type conversationSubscription struct {
	subscribers map[string]chan<- graph.ConversationEvent
	mu          sync.RWMutex
}

// userSubscription represents a subscription to user events.
type userSubscription struct {
	subscribers map[string]chan<- graph.UserEvent
	mu          sync.RWMutex
}

// NewManager creates a new subscription manager.
func NewManager(
	hub *websocket.ChatHub,
	pubSub *pubsub.ChatPubSub,
	logger logger.Logger,
) *Manager {
	m := &Manager{
		logger:         logger,
		subscriptions:  make(map[string]*conversationSubscription),
		userSubs:       make(map[uuid.UUID]*userSubscription),
		Hub:            hub,
		pubSub:         pubSub,
		eventConverter: NewEventConverter(logger),
	}

	// Register handlers for cross-instance messages
	m.registerPubSubHandlers()

	return m
}

// registerPubSubHandlers registers handlers for Redis pub/sub messages.
func (m *Manager) registerPubSubHandlers() {
	// Handle chat messages
	m.pubSub.RegisterHandler(
		pubsub.CrossInstanceMessageTypeChat,
		m.handleCrossInstanceMessage,
	)

	// Handle typing indicators
	m.pubSub.RegisterHandler(
		pubsub.CrossInstanceMessageTypeTyping,
		m.handleCrossInstanceMessage,
	)

	// Handle presence updates
	m.pubSub.RegisterHandler(
		pubsub.CrossInstanceMessageTypePresence,
		m.handleCrossInstanceMessage,
	)

	// Handle delivery receipts
	m.pubSub.RegisterHandler(
		pubsub.CrossInstanceMessageTypeDeliveryReceipt,
		m.handleCrossInstanceMessage,
	)

	// Handle read receipts
	m.pubSub.RegisterHandler(
		pubsub.CrossInstanceMessageTypeReadReceipt,
		m.handleCrossInstanceMessage,
	)
}

// SubscribeToConversation creates a new subscription to conversation events.
func (m *Manager) SubscribeToConversation(
	ctx context.Context,
	conversationID uuid.UUID,
) (<-chan graph.ConversationEvent, error) {
	// Create channel for GraphQL subscription
	ch := make(chan graph.ConversationEvent, constant.SubscriptionChannelBufferSize)
	subID := uuid.New().String()

	m.mu.Lock()

	convSub, exists := m.subscriptions[conversationID.String()]
	if !exists {
		convSub = &conversationSubscription{
			subscribers: make(map[string]chan<- graph.ConversationEvent),
		}
		m.subscriptions[conversationID.String()] = convSub

		// Start listening to ChatHub for this conversation
		go m.listenToConversation(conversationID, convSub)
	}

	m.mu.Unlock()

	// Add this subscriber
	convSub.mu.Lock()
	convSub.subscribers[subID] = ch
	convSub.mu.Unlock()

	// Handle cleanup when context is done
	go func() {
		<-ctx.Done()
		m.unsubscribeFromConversation(conversationID, subID)
		close(ch)
	}()

	return ch, nil
}

// SubscribeToUserEvents creates a new subscription to user events.
func (m *Manager) SubscribeToUserEvents(
	ctx context.Context,
	userID uuid.UUID,
) (<-chan graph.UserEvent, error) {
	// Create channel for GraphQL subscription
	ch := make(chan graph.UserEvent, constant.SubscriptionChannelBufferSize)
	subID := uuid.New().String()

	m.mu.Lock()

	userSub, exists := m.userSubs[userID]
	if !exists {
		userSub = &userSubscription{
			subscribers: make(map[string]chan<- graph.UserEvent),
		}
		m.userSubs[userID] = userSub

		// Start listening to ChatHub for this user
		go m.listenToUser(userID, userSub)
	}

	m.mu.Unlock()

	// Add this subscriber
	userSub.mu.Lock()
	userSub.subscribers[subID] = ch
	userSub.mu.Unlock()

	// Handle cleanup when context is done
	go func() {
		<-ctx.Done()
		m.unsubscribeFromUser(userID, subID)
		close(ch)
	}()

	return ch, nil
}

// handleCrossInstanceMessage handles incoming cross-instance messages from Redis.
func (m *Manager) handleCrossInstanceMessage(
	_ context.Context,
	crossMsg *pubsub.CrossInstanceMessage,
) error {
	// Unmarshal the payload to CrossInstancePayload
	var payload websocket.CrossInstancePayload
	if err := json.Unmarshal(crossMsg.Payload, &payload); err != nil {
		m.logger.Error("Failed to unmarshal cross-instance payload",
			"error", err,
			"message_type", crossMsg.MessageType)

		return err
	}

	// Convert CrossInstancePayload to WebSocket message
	wsMsg := &pkgwebsocket.Message{
		ID:        payload.ID,
		Type:      payload.Type,
		Channel:   payload.Channel,
		SenderID:  payload.SenderID,
		Content:   payload.Content,
		Timestamp: payload.Timestamp,
	}

	// Route message based on type
	switch crossMsg.MessageType {
	case pubsub.CrossInstanceMessageTypePresence:
		return m.handlePresenceMessage(wsMsg)
	case pubsub.CrossInstanceMessageTypeChat,
		pubsub.CrossInstanceMessageTypeTyping,
		pubsub.CrossInstanceMessageTypeDeliveryReceipt,
		pubsub.CrossInstanceMessageTypeReadReceipt:
		if crossMsg.ConversationID != nil {
			return m.handleConversationMessage(wsMsg, *crossMsg.ConversationID)
		}
	}

	return nil
}

// handleConversationMessage broadcasts a message to all conversation subscribers.
func (m *Manager) handleConversationMessage(
	msg *pkgwebsocket.Message,
	conversationID uuid.UUID,
) error {
	// Convert WebSocket message to GraphQL event
	event, err := m.eventConverter.ToConversationEvent(msg)
	if err != nil {
		m.logger.Error("Failed to convert message to GraphQL event",
			"error", err,
			"conversation_id", conversationID)

		return err
	}

	if event == nil {
		// Not a conversation event, skip
		return nil
	}

	// Find subscribers for this conversation
	m.mu.RLock()
	convSub, exists := m.subscriptions[conversationID.String()]
	m.mu.RUnlock()

	if !exists {
		// No subscribers for this conversation
		return nil
	}

	// Broadcast to all subscribers
	convSub.mu.RLock()
	defer convSub.mu.RUnlock()

	for _, sub := range convSub.subscribers {
		select {
		case sub <- event:
		default:
			m.logger.Warn("Subscriber channel full, dropping message",
				"conversation_id", conversationID)
		}
	}

	return nil
}

// handlePresenceMessage broadcasts a presence update to all user subscribers.
func (m *Manager) handlePresenceMessage(msg *pkgwebsocket.Message) error {
	// Convert WebSocket message to GraphQL event
	event, err := m.eventConverter.ToUserEvent(msg)
	if err != nil {
		m.logger.Error("Failed to convert message to GraphQL user event",
			"error", err)

		return err
	}

	if event == nil {
		// Not a user event, skip
		return nil
	}

	// Extract user ID from the presence event
	presenceUpdate, ok := event.(*graph.PresenceUpdate)
	if !ok {
		return nil
	}

	userID, err := uuid.Parse(presenceUpdate.UserID)
	if err != nil {
		m.logger.Error("Failed to parse user ID from presence update",
			"error", err,
			"user_id", presenceUpdate.UserID)

		return err
	}

	// Find subscribers for this user
	m.mu.RLock()
	userSub, exists := m.userSubs[userID]
	m.mu.RUnlock()

	if !exists {
		// No subscribers for this user
		return nil
	}

	// Broadcast to all subscribers
	userSub.mu.RLock()
	defer userSub.mu.RUnlock()

	for _, sub := range userSub.subscribers {
		select {
		case sub <- event:
		default:
			m.logger.Warn("Subscriber channel full, dropping message",
				"user_id", userID)
		}
	}

	return nil
}

// listenToConversation listens to Redis pub/sub events for a conversation.
// Note: This method is now a no-op placeholder since Redis pub/sub handlers
// are registered globally and route messages to subscribers automatically.
func (m *Manager) listenToConversation(
	_ uuid.UUID,
	_ *conversationSubscription,
) {
	// Redis pub/sub handlers registered in registerPubSubHandlers will
	// automatically route messages to the appropriate subscribers via
	// handleCrossInstanceMessage -> handleConversationMessage
	//
	// This method is kept for interface compatibility but does nothing
	// since subscription management is now fully handled by Redis pub/sub.
}

// listenToUser listens to Redis pub/sub events for a user.
// Note: This method is now a no-op placeholder since Redis pub/sub handlers
// are registered globally and route messages to subscribers automatically.
func (m *Manager) listenToUser(_ uuid.UUID, _ *userSubscription) {
	// Redis pub/sub handlers registered in registerPubSubHandlers will
	// automatically route messages to the appropriate subscribers via
	// handleCrossInstanceMessage -> handlePresenceMessage
	//
	// This method is kept for interface compatibility but does nothing
	// since subscription management is now fully handled by Redis pub/sub.
}

// unsubscribeFromConversation removes a subscriber from a conversation.
func (m *Manager) unsubscribeFromConversation(conversationID uuid.UUID, subID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	convSub, exists := m.subscriptions[conversationID.String()]
	if !exists {
		return
	}

	convSub.mu.Lock()
	delete(convSub.subscribers, subID)
	subscriberCount := len(convSub.subscribers)
	convSub.mu.Unlock()

	// If no more subscribers, remove the conversation subscription
	if subscriberCount == 0 {
		delete(m.subscriptions, conversationID.String())
	}
}

// unsubscribeFromUser removes a subscriber from user events.
func (m *Manager) unsubscribeFromUser(userID uuid.UUID, subID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userSub, exists := m.userSubs[userID]
	if !exists {
		return
	}

	userSub.mu.Lock()
	delete(userSub.subscribers, subID)
	subscriberCount := len(userSub.subscribers)
	userSub.mu.Unlock()

	// If no more subscribers, remove the user subscription
	if subscriberCount == 0 {
		delete(m.userSubs, userID)
	}
}
