// Package subscription provides GraphQL subscription infrastructure for bridging EventBus events to GraphQL subscriptions.
package subscription

import (
	"context"

	"github.com/bytedance/sonic"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/rediseventbus"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/event"
)

// EventHandler handles EventBus events and converts them to GraphQL events.
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
func (h *EventHandler) HandleEvent(_ context.Context, evt rediseventbus.Event) error {
	eventType := evt.GetType()

	h.logger.Debug("Received chat event",
		"event_type", eventType,
		"source_instance", evt.GetSourceInstanceID())

	switch eventType {
	case event.TypeChatMessage:
		return h.handleChatMessage(evt)

	case event.TypeTypingIndicator:
		return h.handleTypingIndicator(evt)

	case event.TypePresenceUpdate:
		return h.handlePresenceUpdate(evt)

	case event.TypeDeliveryReceipt:
		return h.handleDeliveryReceipt(evt)

	case event.TypeReadReceipt:
		return h.handleReadReceipt(evt)

	default:
		h.logger.Debug("Unknown chat event type, ignoring",
			"event_type", eventType)

		return nil
	}
}

// handleChatMessage processes chat message events.
func (h *EventHandler) handleChatMessage(evt rediseventbus.Event) error {
	var chatEvent event.ChatMessageEvent

	if err := evt.UnmarshalPayload(&chatEvent); err != nil {
		h.logger.Error("Failed to unmarshal chat message event", "error", err)
		return err
	}

	// Parse WebSocket message content into NewMessage GraphQL type
	var newMsg graph.NewMessage

	if err := sonic.Unmarshal(chatEvent.Message.Content, &newMsg); err != nil {
		h.logger.Error("Failed to unmarshal new message content", "error", err)
		return err
	}

	// Notify GraphQL local subscribers for this conversation
	h.manager.NotifyLocalConversationSubscribers(chatEvent.ConversationID, &newMsg)

	// Broadcast to WebSocket hub if available
	if h.manager.Hub != nil {
		err := h.manager.Hub.BroadcastToConversation(chatEvent.ConversationID, chatEvent.Message)
		if err != nil {
			h.logger.Error("Failed to broadcast chat message event", "error", err)
			return err
		}
	}

	h.logger.Debug("Processed chat message event",
		"conversation_id", chatEvent.ConversationID,
		"message_id", newMsg.ID)

	return nil
}

// handleTypingIndicator processes typing indicator events.
func (h *EventHandler) handleTypingIndicator(evt rediseventbus.Event) error {
	var typingEvent event.TypingIndicatorEvent

	if err := evt.UnmarshalPayload(&typingEvent); err != nil {
		h.logger.Error("Failed to unmarshal typing indicator event", "error", err)
		return err
	}

	// Parse WebSocket message content into TypingIndicator GraphQL type
	var indicator graph.TypingIndicator

	if err := sonic.Unmarshal(typingEvent.Message.Content, &indicator); err != nil {
		h.logger.Error("Failed to unmarshal typing indicator content", "error", err)
		return err
	}

	// Notify GraphQL local subscribers for this conversation
	h.manager.NotifyLocalConversationSubscribers(typingEvent.ConversationID, &indicator)

	// Broadcast to WebSocket hub if available
	if h.manager.Hub != nil {
		err := h.manager.Hub.BroadcastToConversation(
			typingEvent.ConversationID,
			typingEvent.Message,
		)
		if err != nil {
			h.logger.Error("Failed to broadcast typing indicator event", "error", err)
			return err
		}
	}

	h.logger.Debug("Processed typing indicator event",
		"conversation_id", typingEvent.ConversationID,
		"user_id", indicator.UserID)

	return nil
}

// handlePresenceUpdate processes presence update events.
func (h *EventHandler) handlePresenceUpdate(evt rediseventbus.Event) error {
	var presenceEvent event.PresenceUpdateEvent

	if err := evt.UnmarshalPayload(&presenceEvent); err != nil {
		h.logger.Error("Failed to unmarshal presence update event", "error", err)
		return err
	}

	// Parse WebSocket message content into PresenceUpdate GraphQL type
	var update graph.PresenceUpdate

	if err := sonic.Unmarshal(presenceEvent.Message.Content, &update); err != nil {
		h.logger.Error("Failed to unmarshal presence update content", "error", err)
		return err
	}

	// Notify GraphQL local subscribers for this user
	h.manager.NotifyLocalUserSubscribers(presenceEvent.UserID, &update)

	// Broadcast to WebSocket hub if available
	if h.manager.Hub != nil {
		err := h.manager.Hub.BroadcastToUser(presenceEvent.UserID, presenceEvent.Message)
		if err != nil {
			h.logger.Error("Failed to broadcast presence update event", "error", err)
			return err
		}
	}

	h.logger.Debug("Processed presence update event",
		"user_id", presenceEvent.UserID,
		"status", update.Status)

	return nil
}

// handleDeliveryReceipt processes delivery receipt events.
func (h *EventHandler) handleDeliveryReceipt(evt rediseventbus.Event) error {
	var receiptEvent event.DeliveryReceiptEvent

	if err := evt.UnmarshalPayload(&receiptEvent); err != nil {
		h.logger.Error("Failed to unmarshal delivery receipt event", "error", err)
		return err
	}

	// Parse WebSocket message content into DeliveryReceipt GraphQL type
	var receipt graph.DeliveryReceipt

	if err := sonic.Unmarshal(receiptEvent.Message.Content, &receipt); err != nil {
		h.logger.Error("Failed to unmarshal delivery receipt content", "error", err)
		return err
	}

	// Notify GraphQL local subscribers for this conversation
	h.manager.NotifyLocalConversationSubscribers(receiptEvent.ConversationID, &receipt)

	// Broadcast to WebSocket hub if available
	if h.manager.Hub != nil {
		err := h.manager.Hub.BroadcastToConversation(
			receiptEvent.ConversationID,
			receiptEvent.Message,
		)
		if err != nil {
			h.logger.Error("Failed to broadcast delivery receipt event", "error", err)
			return err
		}
	}

	h.logger.Debug("Processed delivery receipt event",
		"conversation_id", receiptEvent.ConversationID,
		"message_id", receipt.MessageID)

	return nil
}

// handleReadReceipt processes read receipt events.
func (h *EventHandler) handleReadReceipt(evt rediseventbus.Event) error {
	var receiptEvent event.ReadReceiptEvent

	if err := evt.UnmarshalPayload(&receiptEvent); err != nil {
		h.logger.Error("Failed to unmarshal read receipt event", "error", err)
		return err
	}

	// Parse WebSocket message content into ReadReceipt GraphQL type
	var receipt graph.ReadReceipt

	if err := sonic.Unmarshal(receiptEvent.Message.Content, &receipt); err != nil {
		h.logger.Error("Failed to unmarshal read receipt content", "error", err)
		return err
	}

	// Notify GraphQL local subscribers for this conversation
	h.manager.NotifyLocalConversationSubscribers(receiptEvent.ConversationID, &receipt)

	// Broadcast to WebSocket hub if available
	if h.manager.Hub != nil {
		err := h.manager.Hub.BroadcastToConversation(
			receiptEvent.ConversationID,
			receiptEvent.Message,
		)
		if err != nil {
			h.logger.Error("Failed to broadcast read receipt event", "error", err)
			return err
		}
	}

	h.logger.Debug("Processed read receipt event",
		"conversation_id", receiptEvent.ConversationID,
		"message_id", receipt.MessageID)

	return nil
}
