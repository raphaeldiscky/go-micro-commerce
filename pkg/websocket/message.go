// Package websocket provides universal WebSocket infrastructure for real-time communication.
package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageType represents the type of WebSocket message.
type MessageType string

// Common message types that can be used across services.
const (
	MessageTypeHeartbeat MessageType = "heartbeat"
	MessageTypeError     MessageType = "error"
	MessageTypeSystem    MessageType = "system"
)

// Message represents a universal WebSocket message envelope.
type Message struct {
	ID        uuid.UUID       `json:"id"`
	Type      MessageType     `json:"type"`
	Channel   *string         `json:"channel,omitempty"`   // Room/channel for broadcasting
	SenderID  *uuid.UUID      `json:"sender_id,omitempty"` // Optional sender identification
	Content   json.RawMessage `json:"content"`             // Flexible content payload
	Timestamp time.Time       `json:"timestamp"`
}

// HeartbeatContent represents heartbeat message content.
type HeartbeatContent struct {
	Ping      bool      `json:"ping,omitempty"`
	Pong      bool      `json:"pong,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorContent represents error message content.
type ErrorContent struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// SystemContent represents system message content.
type SystemContent struct {
	Message string         `json:"message"`
	Event   string         `json:"event"`
	Data    map[string]any `json:"data,omitempty"`
}

// NewMessage creates a new WebSocket message with a generated ID and timestamp.
func NewMessage(msgType MessageType, content any) (*Message, error) {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:        uuid.New(),
		Type:      msgType,
		Content:   contentBytes,
		Timestamp: time.Now(),
	}, nil
}

// NewHeartbeatMessage creates a new heartbeat message.
func NewHeartbeatMessage(ping, pong bool) (*Message, error) {
	content := HeartbeatContent{
		Ping:      ping,
		Pong:      pong,
		Timestamp: time.Now(),
	}

	return NewMessage(MessageTypeHeartbeat, content)
}

// NewErrorMessage creates a new error message.
func NewErrorMessage(code, message string, details any) (*Message, error) {
	content := ErrorContent{
		Code:    code,
		Message: message,
		Details: details,
	}

	return NewMessage(MessageTypeError, content)
}

// NewSystemMessage creates a new system message.
func NewSystemMessage(message, event string, data map[string]any) (*Message, error) {
	content := SystemContent{
		Message: message,
		Event:   event,
		Data:    data,
	}

	return NewMessage(MessageTypeSystem, content)
}

// ParseContent parses the message content into the appropriate struct.
func (m *Message) ParseContent(dest any) error {
	return json.Unmarshal(m.Content, dest)
}

// WithChannel sets the channel for the message.
func (m *Message) WithChannel(channel string) *Message {
	m.Channel = &channel
	return m
}

// WithSender sets the sender ID for the message.
func (m *Message) WithSender(senderID uuid.UUID) *Message {
	m.SenderID = &senderID
	return m
}
