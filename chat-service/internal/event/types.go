// Package event provides chat-specific event types for cross-instance messaging.
package event

import (
	"encoding/json"

	"github.com/google/uuid"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"
)

// Event type constants.
const (
	TypeChatMessage     = "chat_message"
	TypeTypingIndicator = "typing_indicator"
	TypePresenceUpdate  = "presence_update"
	TypeDeliveryReceipt = "delivery_receipt"
	TypeReadReceipt     = "read_receipt"
)

// ChatMessageEvent represents a chat message event for cross-instance delivery.
type ChatMessageEvent struct {
	ConversationID uuid.UUID             `json:"conversation_id"`
	Message        *pkgwebsocket.Message `json:"message"`
	ExcludeUserID  *uuid.UUID            `json:"exclude_user_id,omitempty"`
}

// TypingIndicatorEvent represents a typing indicator event.
type TypingIndicatorEvent struct {
	ConversationID uuid.UUID             `json:"conversation_id"`
	Message        *pkgwebsocket.Message `json:"message"`
	ExcludeUserID  *uuid.UUID            `json:"exclude_user_id,omitempty"`
}

// PresenceUpdateEvent represents a user presence update event.
type PresenceUpdateEvent struct {
	UserID  uuid.UUID             `json:"user_id"`
	Message *pkgwebsocket.Message `json:"message"`
}

// DeliveryReceiptEvent represents a message delivery receipt event.
type DeliveryReceiptEvent struct {
	ConversationID uuid.UUID             `json:"conversation_id"`
	Message        *pkgwebsocket.Message `json:"message"`
	ExcludeUserID  *uuid.UUID            `json:"exclude_user_id,omitempty"`
}

// ReadReceiptEvent represents a message read receipt event.
type ReadReceiptEvent struct {
	ConversationID uuid.UUID             `json:"conversation_id"`
	Message        *pkgwebsocket.Message `json:"message"`
	ExcludeUserID  *uuid.UUID            `json:"exclude_user_id,omitempty"`
}

// UnmarshalChatMessageEvent unmarshals a chat message event.
func UnmarshalChatMessageEvent(data []byte) (*ChatMessageEvent, error) {
	var event ChatMessageEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// UnmarshalTypingIndicatorEvent unmarshals a typing indicator event.
func UnmarshalTypingIndicatorEvent(data []byte) (*TypingIndicatorEvent, error) {
	var event TypingIndicatorEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// UnmarshalPresenceUpdateEvent unmarshals a presence update event.
func UnmarshalPresenceUpdateEvent(data []byte) (*PresenceUpdateEvent, error) {
	var event PresenceUpdateEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// UnmarshalDeliveryReceiptEvent unmarshals a delivery receipt event.
func UnmarshalDeliveryReceiptEvent(data []byte) (*DeliveryReceiptEvent, error) {
	var event DeliveryReceiptEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

// UnmarshalReadReceiptEvent unmarshals a read receipt event.
func UnmarshalReadReceiptEvent(data []byte) (*ReadReceiptEvent, error) {
	var event ReadReceiptEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	return &event, nil
}
