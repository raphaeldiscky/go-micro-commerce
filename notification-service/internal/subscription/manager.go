package subscription

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/eventbus"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
)

// Manager manages GraphQL subscriptions by bridging EventBus events to GraphQL channels.
type Manager struct {
	logger   logger.Logger
	userSubs map[uuid.UUID]*userSubscription
	mu       sync.RWMutex
	EventBus eventbus.EventBus
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
	return &Manager{
		logger:   appLogger,
		userSubs: make(map[uuid.UUID]*userSubscription),
		EventBus: eventBus,
	}
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

	m.mu.Unlock()

	// Add this subscriber
	userSub.mu.Lock()
	userSub.subscribers[subID] = ch
	subscriberCount := len(userSub.subscribers)
	userSub.mu.Unlock()

	m.logger.Info("Added GraphQL notification subscriber",
		"user_id", userID,
		"subscriber_id", subID,
		"total_subscribers", subscriberCount)

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

	// If no more subscribers, remove the user subscription
	if subscriberCount == 0 {
		delete(m.userSubs, userID)

		m.logger.Info("Removed user notification subscription group (no subscribers left)",
			"user_id", userID)
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

	for _, sub := range userSub.subscribers {
		select {
		case sub <- event:
			sentCount++
		default:
			droppedCount++

			m.logger.Warn("Local notification subscriber channel full, dropping message",
				"user_id", userID)
		}
	}

	if sentCount > 0 || droppedCount > 0 {
		m.logger.Debug("Notified local notification subscribers",
			"user_id", userID,
			"subscriber_count", len(userSub.subscribers),
			"sent_count", sentCount,
			"dropped_count", droppedCount)
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
