package subscription

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/rediseventbus"

	redispkg "github.com/raphaeldiscky/go-micro-commerce/pkg/redis"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// Manager manages GraphQL subscriptions by bridging WebSocket events to GraphQL channels.
type Manager struct {
	logger               logger.Logger
	subscriptions        map[string]*conversationSubscription
	userSubs             map[uuid.UUID]*userSubscription
	conversationChannels map[uuid.UUID]string // conversationID → Redis channel name
	userChannels         map[uuid.UUID]string // userID → Redis channel name
	mu                   sync.RWMutex
	Hub                  *websocket.ChatHub
	EventBus             rediseventbus.EventBus
	eventHandlerFunc     rediseventbus.EventHandler
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
	eventBus rediseventbus.EventBus,
	logger logger.Logger,
) *Manager {
	m := &Manager{
		logger:               logger,
		subscriptions:        make(map[string]*conversationSubscription),
		userSubs:             make(map[uuid.UUID]*userSubscription),
		conversationChannels: make(map[uuid.UUID]string),
		userChannels:         make(map[uuid.UUID]string),
		Hub:                  hub,
		EventBus:             eventBus,
	}

	// Create event handler function that routes Redis events to local subscribers
	m.eventHandlerFunc = func(ctx context.Context, event rediseventbus.Event) error {
		return NewEventHandler(m, logger).HandleEvent(ctx, event)
	}

	return m
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

		m.logger.Info("Created new conversation subscription group",
			"conversation_id", conversationID,
			"subscriber_id", subID)
	}

	// Subscribe to Redis sharded channel if this is the first subscriber for this conversation
	isFirstSubscriber := !exists

	m.mu.Unlock()

	// Add this subscriber
	convSub.mu.Lock()
	convSub.subscribers[subID] = ch
	subscriberCount := len(convSub.subscribers)
	convSub.mu.Unlock()

	// Subscribe to Redis if this is the first subscriber for this conversation
	if isFirstSubscriber {
		if err := m.subscribeToConversationRedis(conversationID); err != nil {
			m.logger.Error("Failed to subscribe to Redis for conversation",
				"conversation_id", conversationID,
				"error", err)
			// Don't fail the subscription, local notifications will still work
		}
	}

	m.logger.Info("Added GraphQL subscriber",
		"conversation_id", conversationID,
		"subscriber_id", subID,
		"total_subscribers", subscriberCount,
		"redis_subscribed", isFirstSubscriber)

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

		m.logger.Info("Created new user event subscription group",
			"user_id", userID,
			"subscriber_id", subID)
	}

	// Subscribe to Redis sharded channel if this is the first subscriber for this user
	isFirstSubscriber := !exists

	m.mu.Unlock()

	// Add this subscriber
	userSub.mu.Lock()
	userSub.subscribers[subID] = ch
	subscriberCount := len(userSub.subscribers)
	userSub.mu.Unlock()

	// Subscribe to Redis if this is the first subscriber for this user
	if isFirstSubscriber {
		if err := m.subscribeToUserRedis(userID); err != nil {
			m.logger.Error("Failed to subscribe to Redis for user",
				"user_id", userID,
				"error", err)
			// Don't fail the subscription, local notifications will still work
		}
	}

	m.logger.Info("Added GraphQL user event subscriber",
		"user_id", userID,
		"subscriber_id", subID,
		"total_subscribers", subscriberCount,
		"redis_subscribed", isFirstSubscriber)

	// Handle cleanup when context is done
	go func() {
		<-ctx.Done()
		m.unsubscribeFromUser(userID, subID)
		close(ch)
	}()

	return ch, nil
}

// unsubscribeFromConversation removes a subscriber from a conversation.
func (m *Manager) unsubscribeFromConversation(conversationID uuid.UUID, subID string) {
	m.mu.Lock()

	convSub, exists := m.subscriptions[conversationID.String()]
	if !exists {
		m.mu.Unlock()
		return
	}

	convSub.mu.Lock()
	delete(convSub.subscribers, subID)
	subscriberCount := len(convSub.subscribers)
	convSub.mu.Unlock()

	m.logger.Info("Removed GraphQL subscriber",
		"conversation_id", conversationID,
		"subscriber_id", subID,
		"remaining_subscribers", subscriberCount)

	// If no more subscribers, remove the conversation subscription
	if subscriberCount == 0 {
		delete(m.subscriptions, conversationID.String())
		m.logger.Info("Removed conversation subscription group (no subscribers left)",
			"conversation_id", conversationID)
	}

	// Unlock before calling unsubscribeFromConversationRedis to avoid deadlock
	m.mu.Unlock()

	// Unsubscribe from Redis after releasing the lock
	if subscriberCount == 0 {
		if err := m.unsubscribeFromConversationRedis(conversationID); err != nil {
			m.logger.Error("Failed to unsubscribe from Redis for conversation",
				"conversation_id", conversationID,
				"error", err)
		}
	}
}

// unsubscribeFromUser removes a subscriber from user events.
func (m *Manager) unsubscribeFromUser(userID uuid.UUID, subID string) {
	m.mu.Lock()

	userSub, exists := m.userSubs[userID]
	if !exists {
		m.mu.Unlock()
		return
	}

	userSub.mu.Lock()
	delete(userSub.subscribers, subID)
	subscriberCount := len(userSub.subscribers)
	userSub.mu.Unlock()

	m.logger.Info("Removed GraphQL user event subscriber",
		"user_id", userID,
		"subscriber_id", subID,
		"remaining_subscribers", subscriberCount)

	// If no more subscribers, remove the user subscription
	if subscriberCount == 0 {
		delete(m.userSubs, userID)
		m.logger.Info("Removed user event subscription group (no subscribers left)",
			"user_id", userID)
	}

	// Unlock before calling unsubscribeFromUserRedis to avoid deadlock
	m.mu.Unlock()

	// Unsubscribe from Redis after releasing the lock
	if subscriberCount == 0 {
		if err := m.unsubscribeFromUserRedis(userID); err != nil {
			m.logger.Error("Failed to unsubscribe from Redis for user",
				"user_id", userID,
				"error", err)
		}
	}
}

// NotifyLocalConversationSubscribers directly notifies local GraphQL subscribers for a conversation.
// This is used to notify subscribers on the same instance without going through Redis pub/sub.
func (m *Manager) NotifyLocalConversationSubscribers(
	conversationID uuid.UUID,
	event graph.ConversationEvent,
) {
	m.mu.RLock()
	convSub, exists := m.subscriptions[conversationID.String()]
	m.mu.RUnlock()

	if !exists {
		m.logger.Debug("No local subscribers found for conversation",
			"conversation_id", conversationID,
			"event_type", getEventTypeName(event))

		return
	}

	convSub.mu.RLock()
	defer convSub.mu.RUnlock()

	m.logger.Info("Notifying local GraphQL subscribers",
		"conversation_id", conversationID,
		"event_type", getEventTypeName(event),
		"subscriber_count", len(convSub.subscribers))

	sentCount := 0
	droppedCount := 0

	for _, sub := range convSub.subscribers {
		select {
		case sub <- event:
			sentCount++
		default:
			droppedCount++

			m.logger.Warn("Local subscriber channel full, dropping message",
				"conversation_id", conversationID)
		}
	}

	m.logger.Debug("Local notification completed",
		"conversation_id", conversationID,
		"sent_count", sentCount,
		"dropped_count", droppedCount)
}

// NotifyLocalUserSubscribers directly notifies local GraphQL subscribers for user events.
// This is used to notify subscribers on the same instance without going through Redis pub/sub.
func (m *Manager) NotifyLocalUserSubscribers(userID uuid.UUID, event graph.UserEvent) {
	m.mu.RLock()
	userSub, exists := m.userSubs[userID]
	m.mu.RUnlock()

	if !exists {
		m.logger.Debug("No local user subscribers found",
			"user_id", userID)

		return
	}

	userSub.mu.RLock()
	defer userSub.mu.RUnlock()

	for _, sub := range userSub.subscribers {
		select {
		case sub <- event:
		default:
			m.logger.Warn("Local user subscriber channel full, dropping message",
				"user_id", userID)
		}
	}

	m.logger.Debug("Notified local user subscribers",
		"user_id", userID,
		"subscriber_count", len(userSub.subscribers))
}

// getEventTypeName extracts the type name from a GraphQL event interface for logging.
func getEventTypeName(event any) string {
	if event == nil {
		return "nil"
	}

	t := reflect.TypeOf(event)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

// subscribeToConversationRedis subscribes to the Redis sharded channel for a conversation.
func (m *Manager) subscribeToConversationRedis(conversationID uuid.UUID) error {
	channel := redispkg.ConversationChannel(conversationID)

	m.mu.Lock()
	m.conversationChannels[conversationID] = channel
	m.mu.Unlock()

	// Subscribe to Redis sharded channel
	if err := m.EventBus.SSubscribe(channel, m.eventHandlerFunc); err != nil {
		m.mu.Lock()
		delete(m.conversationChannels, conversationID)
		m.mu.Unlock()

		return err
	}

	m.logger.Info("Subscribed to Redis sharded channel for conversation",
		"conversation_id", conversationID,
		"channel", channel)

	return nil
}

// unsubscribeFromConversationRedis unsubscribes from the Redis sharded channel for a conversation.
func (m *Manager) unsubscribeFromConversationRedis(conversationID uuid.UUID) error {
	m.mu.Lock()

	channel, exists := m.conversationChannels[conversationID]
	if !exists {
		m.mu.Unlock()
		return nil
	}

	delete(m.conversationChannels, conversationID)
	m.mu.Unlock()

	// Unsubscribe from Redis sharded channel
	if err := m.EventBus.SUnsubscribe(channel); err != nil {
		return err
	}

	m.logger.Info("Unsubscribed from Redis sharded channel for conversation",
		"conversation_id", conversationID,
		"channel", channel)

	return nil
}

// subscribeToUserRedis subscribes to the Redis sharded channel for a user.
func (m *Manager) subscribeToUserRedis(userID uuid.UUID) error {
	channel := redispkg.UserPresenceChannel(userID)

	m.mu.Lock()
	m.userChannels[userID] = channel
	m.mu.Unlock()

	// Subscribe to Redis sharded channel
	if err := m.EventBus.SSubscribe(channel, m.eventHandlerFunc); err != nil {
		m.mu.Lock()
		delete(m.userChannels, userID)
		m.mu.Unlock()

		return err
	}

	m.logger.Info("Subscribed to Redis sharded channel for user",
		"user_id", userID,
		"channel", channel)

	return nil
}

// unsubscribeFromUserRedis unsubscribes from the Redis sharded channel for a user.
func (m *Manager) unsubscribeFromUserRedis(userID uuid.UUID) error {
	m.mu.Lock()

	channel, exists := m.userChannels[userID]
	if !exists {
		m.mu.Unlock()
		return nil
	}

	delete(m.userChannels, userID)
	m.mu.Unlock()

	// Unsubscribe from Redis sharded channel
	if err := m.EventBus.SUnsubscribe(channel); err != nil {
		return err
	}

	m.logger.Info("Unsubscribed from Redis sharded channel for user",
		"user_id", userID,
		"channel", channel)

	return nil
}
