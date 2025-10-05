package subscription

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// Manager manages GraphQL subscriptions by bridging WebSocket events to GraphQL channels.
type Manager struct {
	logger         logger.Logger
	subscriptions  map[string]*conversationSubscription
	userSubs       map[uuid.UUID]*userSubscription
	mu             sync.RWMutex
	hub            *websocket.ChatHub
	eventConverter *EventConverter
}

// conversationSubscription represents a subscription to conversation events.
type conversationSubscription struct {
	conversationID uuid.UUID
	subscribers    map[string]chan<- graph.ConversationEvent
	mu             sync.RWMutex
}

// userSubscription represents a subscription to user events.
type userSubscription struct {
	userID      uuid.UUID
	subscribers map[string]chan<- graph.UserEvent
	mu          sync.RWMutex
}

// NewManager creates a new subscription manager.
func NewManager(hub *websocket.ChatHub, logger logger.Logger) *Manager {
	return &Manager{
		logger:         logger,
		subscriptions:  make(map[string]*conversationSubscription),
		userSubs:       make(map[uuid.UUID]*userSubscription),
		hub:            hub,
		eventConverter: NewEventConverter(),
	}
}

// SubscribeToConversation creates a new subscription to conversation events.
func (m *Manager) SubscribeToConversation(
	ctx context.Context,
	conversationID uuid.UUID,
) (<-chan graph.ConversationEvent, error) {
	// Create channel for GraphQL subscription
	ch := make(chan graph.ConversationEvent, 10)
	subID := uuid.New().String()

	m.mu.Lock()
	convSub, exists := m.subscriptions[conversationID.String()]
	if !exists {
		convSub = &conversationSubscription{
			conversationID: conversationID,
			subscribers:    make(map[string]chan<- graph.ConversationEvent),
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
	ch := make(chan graph.UserEvent, 10)
	subID := uuid.New().String()

	m.mu.Lock()
	userSub, exists := m.userSubs[userID]
	if !exists {
		userSub = &userSubscription{
			userID:      userID,
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

// listenToConversation listens to ChatHub events for a conversation and forwards to subscribers.
func (m *Manager) listenToConversation(conversationID uuid.UUID, convSub *conversationSubscription) {
	// Subscribe to ChatHub conversation channel
	channelName := websocket.ConversationChannel(conversationID)
	messageChan := make(chan *pkgwebsocket.Message, 100)

	// Register this channel with the hub to receive messages
	// Note: This requires adding SubscribeToChannel method to BaseHub
	unsubscribe := m.hub.SubscribeToChannel(channelName, messageChan)
	defer unsubscribe()

	for msg := range messageChan {
		// Convert WebSocket message to GraphQL event
		event, err := m.eventConverter.ToConversationEvent(msg)
		if err != nil {
			m.logger.Error("Failed to convert message to GraphQL event",
				"error", err,
				"conversation_id", conversationID)
			continue
		}

		if event == nil {
			// Not a conversation event, skip
			continue
		}

		// Broadcast to all subscribers
		convSub.mu.RLock()
		for _, sub := range convSub.subscribers {
			select {
			case sub <- event:
			default:
				m.logger.Warn("Subscriber channel full, dropping message")
			}
		}
		convSub.mu.RUnlock()
	}
}

// listenToUser listens to ChatHub events for a user and forwards to subscribers.
func (m *Manager) listenToUser(userID uuid.UUID, userSub *userSubscription) {
	// Subscribe to ChatHub user channel
	channelName := websocket.UserChannel(userID)
	messageChan := make(chan *pkgwebsocket.Message, 100)

	// Register this channel with the hub
	unsubscribe := m.hub.SubscribeToChannel(channelName, messageChan)
	defer unsubscribe()

	for msg := range messageChan {
		// Convert WebSocket message to GraphQL event
		event, err := m.eventConverter.ToUserEvent(msg)
		if err != nil {
			m.logger.Error("Failed to convert message to GraphQL user event",
				"error", err,
				"user_id", userID)
			continue
		}

		if event == nil {
			// Not a user event, skip
			continue
		}

		// Broadcast to all subscribers
		userSub.mu.RLock()
		for _, sub := range userSub.subscribers {
			select {
			case sub <- event:
			default:
				m.logger.Warn("Subscriber channel full, dropping message")
			}
		}
		userSub.mu.RUnlock()
	}
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
