package subscription

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/eventbus"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
)

// Manager manages GraphQL subscriptions by bridging EventBus events to GraphQL channels.
type Manager struct {
	logger           logger.Logger
	userSubs         map[uuid.UUID]*userSubscription
	redisChannels    map[uuid.UUID]string // userID → Redis channel name
	mu               sync.RWMutex
	EventBus         eventbus.EventBus
	eventHandlerFunc eventbus.EventHandler
}

// userSubscription represents a subscription to user notification events.
type userSubscription struct {
	subscribers map[string]chan<- graph.NotificationEvent
	mu          sync.RWMutex
}

// NewManager creates a new subscription manager.
func NewManager(
	eventBus eventbus.EventBus,
	appLogger logger.Logger,
) *Manager {
	m := &Manager{
		logger:        appLogger,
		userSubs:      make(map[uuid.UUID]*userSubscription),
		redisChannels: make(map[uuid.UUID]string),
		EventBus:      eventBus,
	}

	// Create event handler function that routes Redis events to local subscribers
	m.eventHandlerFunc = func(ctx context.Context, event eventbus.Event) error {
		return NewEventHandler(m, appLogger).HandleEvent(ctx, event)
	}

	return m
}

// SubscribeToNotifications creates a new subscription to notification events for a user.
func (m *Manager) SubscribeToNotifications(
	ctx context.Context,
	userID uuid.UUID,
) (<-chan graph.NotificationEvent, error) {
	// Create channel for GraphQL subscription
	ch := make(chan graph.NotificationEvent, constant.SubscriptionChannelBufferSize)
	subID := uuid.New().String()

	m.mu.Lock()

	userSub, exists := m.userSubs[userID]
	if !exists {
		userSub = &userSubscription{
			subscribers: make(map[string]chan<- graph.NotificationEvent),
		}
		m.userSubs[userID] = userSub

		m.logger.Info("Created new user notification subscription group",
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
		if err := m.subscribeToRedis(userID); err != nil {
			m.logger.Error("Failed to subscribe to Redis for user",
				"user_id", userID,
				"error", err)
			// Don't fail the subscription, local notifications will still work
		}
	}

	m.logger.Info("Added GraphQL notification subscriber",
		"user_id", userID,
		"subscriber_id", subID,
		"total_subscribers", subscriberCount,
		"redis_subscribed", isFirstSubscriber)

	// Handle cleanup when context is done
	go func() {
		<-ctx.Done()
		m.unsubscribeFromUser(userID, subID)
		close(ch)

		m.logger.Info("GraphQL notification subscription closed",
			"user_id", userID,
			"subscriber_id", subID)
	}()

	return ch, nil
}

// unsubscribeFromUser removes a subscriber from user notification events.
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

	m.logger.Info("Removed GraphQL notification subscriber",
		"user_id", userID,
		"subscriber_id", subID,
		"remaining_subscribers", subscriberCount)

	// If no more subscribers, remove the user subscription and unsubscribe from Redis
	if subscriberCount == 0 {
		delete(m.userSubs, userID)

		m.logger.Info("Removed user notification subscription group (no subscribers left)",
			"user_id", userID)

		// Unsubscribe from Redis
		if err := m.unsubscribeFromRedis(userID); err != nil {
			m.logger.Error("Failed to unsubscribe from Redis for user",
				"user_id", userID,
				"error", err)
		}
	}
}

// NotifyLocalSubscribers directly notifies local GraphQL subscribers for user notification events.
// This is used to notify subscribers on the same instance without going through Redis pub/sub.
func (m *Manager) NotifyLocalSubscribers(userID uuid.UUID, event graph.NotificationEvent) {
	m.mu.RLock()
	userSub, exists := m.userSubs[userID]
	m.mu.RUnlock()

	if !exists {
		m.logger.Debug("No local notification subscribers found",
			"user_id", userID)

		return
	}

	userSub.mu.RLock()
	defer userSub.mu.RUnlock()

	sentCount := 0
	droppedCount := 0

	for subID, sub := range userSub.subscribers {
		select {
		case sub <- event:
			sentCount++

			m.logger.Debug("Event sent to GraphQL subscriber channel",
				"user_id", userID,
				"subscriber_id", subID,
				"event_type", getEventTypeName(event))
		default:
			droppedCount++

			m.logger.Warn("Local notification subscriber channel full, dropping message",
				"user_id", userID,
				"subscriber_id", subID)
		}
	}

	if sentCount > 0 || droppedCount > 0 {
		m.logger.Info("Notified local notification subscribers",
			"user_id", userID,
			"subscriber_count", len(userSub.subscribers),
			"sent_count", sentCount,
			"dropped_count", droppedCount,
			"event_type", getEventTypeName(event))
	}
}

// GetSubscriberCount returns the number of active subscribers for a user.
func (m *Manager) GetSubscriberCount(userID uuid.UUID) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userSub, exists := m.userSubs[userID]
	if !exists {
		return 0
	}

	userSub.mu.RLock()
	defer userSub.mu.RUnlock()

	return len(userSub.subscribers)
}

// getEventTypeName returns a human-readable name for the event type.
func getEventTypeName(event graph.NotificationEvent) string {
	switch event.(type) {
	case *graph.NewNotification:
		return "NewNotification"
	case *graph.NotificationRead:
		return "NotificationRead"
	case *graph.NotificationDeleted:
		return "NotificationDeleted"
	default:
		return "Unknown"
	}
}

// GetAllSubscriptions returns a map of all active subscriptions by user ID.
func (m *Manager) GetAllSubscriptions() map[uuid.UUID]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[uuid.UUID]int)

	for userID, userSub := range m.userSubs {
		userSub.mu.RLock()
		result[userID] = len(userSub.subscribers)
		userSub.mu.RUnlock()
	}

	return result
}

// subscribeToRedis subscribes to the Redis sharded channel for a user.
func (m *Manager) subscribeToRedis(userID uuid.UUID) error {
	channel := redis.NotificationUserChannel(userID)

	m.mu.Lock()
	m.redisChannels[userID] = channel
	m.mu.Unlock()

	// Subscribe to Redis sharded channel
	if err := m.EventBus.SSubscribe(channel, m.eventHandlerFunc); err != nil {
		m.mu.Lock()
		delete(m.redisChannels, userID)
		m.mu.Unlock()

		return err
	}

	m.logger.Info("Subscribed to Redis sharded channel",
		"user_id", userID,
		"channel", channel)

	return nil
}

// unsubscribeFromRedis unsubscribes from the Redis sharded channel for a user.
func (m *Manager) unsubscribeFromRedis(userID uuid.UUID) error {
	m.mu.Lock()
	channel, exists := m.redisChannels[userID]
	if !exists {
		m.mu.Unlock()
		return nil
	}
	delete(m.redisChannels, userID)
	m.mu.Unlock()

	// Unsubscribe from Redis sharded channel
	if err := m.EventBus.SUnsubscribe(channel); err != nil {
		return err
	}

	m.logger.Info("Unsubscribed from Redis sharded channel",
		"user_id", userID,
		"channel", channel)

	return nil
}
