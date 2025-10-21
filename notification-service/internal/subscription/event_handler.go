// Package subscription provides GraphQL subscription infrastructure for bridging webSocket events to GraphQL subscriptions over graphql-transport-ws protocol.
package subscription

import (
	"context"
	"encoding/json"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/rediseventbus"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/mapper"
)

// EventHandler handles EventBus notifications and converts them to GraphQL events.
type EventHandler struct {
	manager *Manager
	logger  logger.Logger
}

// NewEventHandler creates a new event handler.
func NewEventHandler(
	manager *Manager,
	appLogger logger.Logger,
) *EventHandler {
	return &EventHandler{
		manager: manager,
		logger:  appLogger,
	}
}

// HandleEvent processes events from the EventBus and notifies GraphQL subscribers.
func (h *EventHandler) HandleEvent(_ context.Context, event rediseventbus.Event) error {
	eventType := event.GetType()

	h.logger.Debug("Received notification event",
		"event_type", eventType,
		"source_instance", event.GetSourceInstanceID())

	switch eventType {
	case TypeNotificationCreated:
		return h.handleNotificationCreated(event)

	case TypeNotificationRead:
		return h.handleNotificationRead(event)

	case TypeNotificationDeleted:
		return h.handleNotificationDeleted(event)

	default:
		h.logger.Debug("Unknown notification event type, ignoring",
			"event_type", eventType)

		return nil
	}
}

// handleNotificationCreated processes notification created events.
func (h *EventHandler) handleNotificationCreated(event rediseventbus.Event) error {
	// Parse the notification created event from payload
	var createdEvent NotificationCreatedEvent

	if err := event.UnmarshalPayload(&createdEvent); err != nil {
		h.logger.Error("Failed to unmarshal notification created event", "error", err)
		return err
	}

	// Extract notification DTO from the SSE message
	var notifDTO dto.NotificationResponse

	if err := json.Unmarshal(createdEvent.Message.Data, &notifDTO); err != nil {
		h.logger.Error("Failed to unmarshal notification data", "error", err)
		return err
	}

	// Convert to GraphQL NewNotification event using mapper
	graphQLEvent := mapper.MapToGraphQLNewNotificationFromDTO(&notifDTO)

	// Notify GraphQL local subscribers
	h.manager.NotifyLocalSubscribers(createdEvent.UserID, graphQLEvent)

	// Broadcast to SSE connections if SSE hub is available
	if h.manager.sseHub != nil {
		if err := h.manager.sseHub.BroadcastToUser(createdEvent.UserID, createdEvent.Message); err != nil {
			h.logger.Warn("Failed to broadcast to SSE connections",
				"user_id", createdEvent.UserID,
				"error", err)
		}
	}

	h.logger.Debug("Processed notification created event",
		"user_id", createdEvent.UserID,
		"notification_id", graphQLEvent.ID)

	return nil
}

// handleNotificationRead processes notification read events.
func (h *EventHandler) handleNotificationRead(event rediseventbus.Event) error {
	var readEvent NotificationReadEvent

	if err := event.UnmarshalPayload(&readEvent); err != nil {
		h.logger.Error("Failed to unmarshal notification read event", "error", err)
		return err
	}

	// Convert to GraphQL NotificationRead event
	graphQLEvent := &graph.NotificationRead{
		ID:     readEvent.NotificationID,
		UserID: readEvent.UserID,
		ReadAt: time.Now(),
	}

	// Notify local subscribers
	h.manager.NotifyLocalSubscribers(readEvent.UserID, graphQLEvent)

	h.logger.Debug("Processed notification read event",
		"user_id", readEvent.UserID,
		"notification_id", readEvent.NotificationID)

	return nil
}

// handleNotificationDeleted processes notification deleted events.
func (h *EventHandler) handleNotificationDeleted(event rediseventbus.Event) error {
	var deletedEvent NotificationDeletedEvent

	if err := event.UnmarshalPayload(&deletedEvent); err != nil {
		h.logger.Error("Failed to unmarshal notification deleted event", "error", err)
		return err
	}

	// Convert to GraphQL NotificationDeleted event
	graphQLEvent := &graph.NotificationDeleted{
		ID:     deletedEvent.NotificationID,
		UserID: deletedEvent.UserID,
	}

	// Notify local subscribers
	h.manager.NotifyLocalSubscribers(deletedEvent.UserID, graphQLEvent)

	h.logger.Debug("Processed notification deleted event",
		"user_id", deletedEvent.UserID,
		"notification_id", deletedEvent.NotificationID)

	return nil
}
