package notification

import (
	"context"
	"fmt"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/eventbus"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

// EventHandler handles notification-related events from other instances.
type EventHandler struct {
	logger              logger.Logger
	notificationCreated func(ctx context.Context, event *CreatedEvent) error
	notificationRead    func(ctx context.Context, event *ReadEvent) error
	notificationDeleted func(ctx context.Context, event *DeletedEvent) error
}

// NewEventHandler creates a new notification event handler.
func NewEventHandler(logger logger.Logger) *EventHandler {
	return &EventHandler{
		logger: logger,
	}
}

// SetNotificationCreatedHandler sets the handler for notification created events.
func (h *EventHandler) SetNotificationCreatedHandler(
	handler func(ctx context.Context, event *CreatedEvent) error,
) {
	h.notificationCreated = handler
}

// SetNotificationReadHandler sets the handler for notification read events.
func (h *EventHandler) SetNotificationReadHandler(
	handler func(ctx context.Context, event *ReadEvent) error,
) {
	h.notificationRead = handler
}

// SetNotificationDeletedHandler sets the handler for notification deleted events.
func (h *EventHandler) SetNotificationDeletedHandler(
	handler func(ctx context.Context, event *DeletedEvent) error,
) {
	h.notificationDeleted = handler
}

// eventHandler defines a generic event handler function signature.
type eventHandler[T any] func(ctx context.Context, event T) error

// handleEventType is a generic helper that handles the common pattern of:
// 1. Checking if handler is registered
// 2. Unmarshaling the event payload
// 3. Calling the handler function.
func handleEventType[T any](
	ctx context.Context,
	logger logger.Logger,
	eventType string,
	payload []byte,
	handler eventHandler[T],
	unmarshalFunc func([]byte) (T, error),
) error {
	if handler == nil {
		logger.Warn("No handler registered", "event_type", eventType)
		return nil
	}

	event, err := unmarshalFunc(payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %s event: %w", eventType, err)
	}

	return handler(ctx, event)
}

// HandleEvent is the main event handler that routes events to specific handlers.
func (h *EventHandler) HandleEvent(ctx context.Context, event eventbus.Event) error {
	baseEvent, ok := event.(*eventbus.BaseEvent)
	if !ok {
		return fmt.Errorf("invalid event type: %T", event)
	}

	switch baseEvent.EventType {
	case TypeNotificationCreated:
		return handleEventType(ctx, h.logger, "notification created", baseEvent.Payload,
			h.notificationCreated, UnmarshalCreatedEvent)

	case TypeNotificationRead:
		return handleEventType(ctx, h.logger, "notification read", baseEvent.Payload,
			h.notificationRead, UnmarshalReadEvent)

	case TypeNotificationDeleted:
		return handleEventType(ctx, h.logger, "notification deleted", baseEvent.Payload,
			h.notificationDeleted, UnmarshalDeletedEvent)

	default:
		h.logger.Warn("Unknown event type", "type", baseEvent.EventType)
		return nil
	}
}
