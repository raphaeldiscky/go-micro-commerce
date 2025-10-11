package event

import (
	"context"
	"fmt"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/eventbus"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
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

// HandleEvent is the main event handler that routes events to specific handlers.
func (h *ChatEventHandler) HandleEvent(ctx context.Context, event eventbus.Event) error {
	baseEvent, ok := event.(*eventbus.BaseEvent)
	if !ok {
		return fmt.Errorf("invalid event type: %T", event)
	}

	switch baseEvent.EventType {
	case TypeChatMessage:
		if h.chatMessageFunc == nil {
			h.logger.Warn("No handler registered for chat message events")
			return nil
		}

		chatEvent, err := UnmarshalChatMessageEvent(baseEvent.Payload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal chat message event: %w", err)
		}

		return h.chatMessageFunc(ctx, chatEvent)

	case TypeTypingIndicator:
		if h.typingFunc == nil {
			h.logger.Warn("No handler registered for typing indicator events")
			return nil
		}

		typingEvent, err := UnmarshalTypingIndicatorEvent(baseEvent.Payload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal typing indicator event: %w", err)
		}

		return h.typingFunc(ctx, typingEvent)

	case TypePresenceUpdate:
		if h.presenceFunc == nil {
			h.logger.Warn("No handler registered for presence update events")
			return nil
		}

		presenceEvent, err := UnmarshalPresenceUpdateEvent(baseEvent.Payload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal presence update event: %w", err)
		}

		return h.presenceFunc(ctx, presenceEvent)

	case TypeDeliveryReceipt:
		if h.deliveryReceiptFunc == nil {
			h.logger.Warn("No handler registered for delivery receipt events")
			return nil
		}

		deliveryEvent, err := UnmarshalDeliveryReceiptEvent(baseEvent.Payload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal delivery receipt event: %w", err)
		}

		return h.deliveryReceiptFunc(ctx, deliveryEvent)

	case TypeReadReceipt:
		if h.readReceiptFunc == nil {
			h.logger.Warn("No handler registered for read receipt events")
			return nil
		}

		readEvent, err := UnmarshalReadReceiptEvent(baseEvent.Payload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal read receipt event: %w", err)
		}

		return h.readReceiptFunc(ctx, readEvent)

	default:
		h.logger.Warn("Unknown event type", "type", baseEvent.EventType)
		return nil
	}
}
