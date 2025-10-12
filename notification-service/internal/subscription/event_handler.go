// Package subscription provides GraphQL subscription infrastructure for bridging webSocket events to GraphQL subscriptions over graphql-transport-ws protocol.
package subscription

import (
	"context"
	"encoding/json"
	"time"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/eventbus"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/notification"
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
func (h *EventHandler) HandleEvent(_ context.Context, event eventbus.Event) error {
	eventType := event.GetType()

	h.logger.Debug("Received notification event",
		"event_type", eventType,
		"source_instance", event.GetSourceInstanceID())

	switch eventType {
	case notification.TypeNotificationCreated:
		return h.handleNotificationCreated(event)

	case notification.TypeNotificationRead:
		return h.handleNotificationRead(event)

	case notification.TypeNotificationDeleted:
		return h.handleNotificationDeleted(event)

	default:
		h.logger.Debug("Unknown notification event type, ignoring",
			"event_type", eventType)

		return nil
	}
}

// handleNotificationCreated processes notification created events.
func (h *EventHandler) handleNotificationCreated(event eventbus.Event) error {
	// Marshal event to get raw data
	data, err := event.Marshal()
	if err != nil {
		h.logger.Error("Failed to marshal notification created event", "error", err)
		return err
	}

	// Parse the notification created event
	var createdEvent notification.CreatedEvent

	if err = json.Unmarshal(data, &createdEvent); err != nil {
		h.logger.Error("Failed to unmarshal notification created event", "error", err)
		return err
	}

	// Extract notification data from the SSE message
	var notifData map[string]interface{}

	if err = json.Unmarshal(createdEvent.Message.Data, &notifData); err != nil {
		h.logger.Error("Failed to unmarshal notification data", "error", err)
		return err
	}

	// Convert to GraphQL NewNotification event
	graphQLEvent := &graph.NewNotification{
		ID:        getString(notifData, "id"),
		UserID:    createdEvent.UserID.String(),
		Type:      getNotificationType(notifData),
		Title:     getString(notifData, "title"),
		Message:   getString(notifData, "message"),
		Metadata:  getStringPtr(notifData, "metadata"),
		IsRead:    false,
		CreatedAt: createdEvent.Message.CreatedAt,
	}

	// Notify local subscribers
	h.manager.NotifyLocalSubscribers(createdEvent.UserID, graphQLEvent)

	h.logger.Debug("Processed notification created event",
		"user_id", createdEvent.UserID,
		"notification_id", graphQLEvent.ID)

	return nil
}

// handleNotificationRead processes notification read events.
func (h *EventHandler) handleNotificationRead(event eventbus.Event) error {
	data, err := event.Marshal()
	if err != nil {
		h.logger.Error("Failed to marshal notification read event", "error", err)
		return err
	}

	var readEvent notification.ReadEvent

	if err = json.Unmarshal(data, &readEvent); err != nil {
		h.logger.Error("Failed to unmarshal notification read event", "error", err)
		return err
	}

	// Convert to GraphQL NotificationRead event
	graphQLEvent := &graph.NotificationRead{
		ID:     readEvent.NotificationID.String(),
		UserID: readEvent.UserID.String(),
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
func (h *EventHandler) handleNotificationDeleted(event eventbus.Event) error {
	data, err := event.Marshal()
	if err != nil {
		h.logger.Error("Failed to marshal notification deleted event", "error", err)
		return err
	}

	var deletedEvent notification.DeletedEvent

	if err = json.Unmarshal(data, &deletedEvent); err != nil {
		h.logger.Error("Failed to unmarshal notification deleted event", "error", err)
		return err
	}

	// Convert to GraphQL NotificationDeleted event
	graphQLEvent := &graph.NotificationDeleted{
		ID:     deletedEvent.NotificationID.String(),
		UserID: deletedEvent.UserID.String(),
	}

	// Notify local subscribers
	h.manager.NotifyLocalSubscribers(deletedEvent.UserID, graphQLEvent)

	h.logger.Debug("Processed notification deleted event",
		"user_id", deletedEvent.UserID,
		"notification_id", deletedEvent.NotificationID)

	return nil
}

// Helper functions to safely extract data from notification payload.

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}

	return ""
}

func getStringPtr(data map[string]interface{}, key string) *string {
	if val, ok := data[key].(string); ok {
		return &val
	}

	return nil
}

func getNotificationType(data map[string]interface{}) constant.NotificationType {
	typeStr := getString(data, "type")

	return constant.NotificationType(typeStr)
}
