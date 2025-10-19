package event

import (
	"context"
	"fmt"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/rediseventbus"
)

// ChatEventHandler handles chat-related events from other instances.
type ChatEventHandler struct {
	logger              logger.Logger
	chatMessageFunc     func(ctx context.Context, event *ChatMessageEvent) error
	typingFunc          func(ctx context.Context, event *TypingIndicatorEvent) error
	presenceFunc        func(ctx context.Context, event *PresenceUpdateEvent) error
	deliveryReceiptFunc func(ctx context.Context, event *DeliveryReceiptEvent) error
	readReceiptFunc     func(ctx context.Context, event *ReadReceiptEvent) error
}

// NewChatEventHandler creates a new chat event handler.
func NewChatEventHandler(logger logger.Logger) *ChatEventHandler {
	return &ChatEventHandler{
		logger: logger,
	}
}

// SetChatMessageHandler sets the handler for chat message events.
func (h *ChatEventHandler) SetChatMessageHandler(
	handler func(ctx context.Context, event *ChatMessageEvent) error,
) {
	h.chatMessageFunc = handler
}

// SetTypingIndicatorHandler sets the handler for typing indicator events.
func (h *ChatEventHandler) SetTypingIndicatorHandler(
	handler func(ctx context.Context, event *TypingIndicatorEvent) error,
) {
	h.typingFunc = handler
}

// SetPresenceUpdateHandler sets the handler for presence update events.
func (h *ChatEventHandler) SetPresenceUpdateHandler(
	handler func(ctx context.Context, event *PresenceUpdateEvent) error,
) {
	h.presenceFunc = handler
}

// SetDeliveryReceiptHandler sets the handler for delivery receipt events.
func (h *ChatEventHandler) SetDeliveryReceiptHandler(
	handler func(ctx context.Context, event *DeliveryReceiptEvent) error,
) {
	h.deliveryReceiptFunc = handler
}

// SetReadReceiptHandler sets the handler for read receipt events.
func (h *ChatEventHandler) SetReadReceiptHandler(
	handler func(ctx context.Context, event *ReadReceiptEvent) error,
) {
	h.readReceiptFunc = handler
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
func (h *ChatEventHandler) HandleEvent(ctx context.Context, event rediseventbus.Event) error {
	baseEvent, ok := event.(*rediseventbus.BaseEvent)
	if !ok {
		return fmt.Errorf("invalid event type: %T", event)
	}

	switch baseEvent.EventType {
	case TypeChatMessage:
		return handleEventType(ctx, h.logger, "chat message", baseEvent.Payload,
			h.chatMessageFunc, UnmarshalChatMessageEvent)

	case TypeTypingIndicator:
		return handleEventType(ctx, h.logger, "typing indicator", baseEvent.Payload,
			h.typingFunc, UnmarshalTypingIndicatorEvent)

	case TypePresenceUpdate:
		return handleEventType(ctx, h.logger, "presence update", baseEvent.Payload,
			h.presenceFunc, UnmarshalPresenceUpdateEvent)

	case TypeDeliveryReceipt:
		return handleEventType(ctx, h.logger, "delivery receipt", baseEvent.Payload,
			h.deliveryReceiptFunc, UnmarshalDeliveryReceiptEvent)

	case TypeReadReceipt:
		return handleEventType(ctx, h.logger, "read receipt", baseEvent.Payload,
			h.readReceiptFunc, UnmarshalReadReceiptEvent)

	default:
		h.logger.Warn("Unknown event type", "type", baseEvent.EventType)
		return nil
	}
}
