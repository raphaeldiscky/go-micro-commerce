package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"
)

// MessageParser handles parsing of WebSocket messages from cross-instance payloads.
type MessageParser interface {
	ParseWebSocketMessage(payload json.RawMessage) (*pkgwebsocket.Message, error)
}

type messageParser struct{}

// NewMessageParser creates a new message parser.
func NewMessageParser() MessageParser {
	return &messageParser{}
}

// messageData represents the structure of cross-instance message data.
type messageData struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Channel   *string     `json:"channel,omitempty"`
	SenderID  *string     `json:"sender_id,omitempty"`
	Content   interface{} `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
}

// ParseWebSocketMessage parses a WebSocket message from cross-instance payload.
func (mp *messageParser) ParseWebSocketMessage(
	payload json.RawMessage,
) (*pkgwebsocket.Message, error) {
	var data messageData
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}

	msg := &pkgwebsocket.Message{
		Type:      pkgwebsocket.MessageType(data.Type),
		Channel:   data.Channel,
		Timestamp: data.Timestamp,
	}

	// Parse ID
	if data.ID != "" {
		if parsedID, err := uuid.Parse(data.ID); err == nil {
			msg.ID = parsedID
		}
	}

	// Parse SenderID
	if data.SenderID != nil && *data.SenderID != "" {
		if parsedID, err := uuid.Parse(*data.SenderID); err == nil {
			msg.SenderID = &parsedID
		}
	}

	// Parse Content
	if data.Content != nil {
		if contentBytes, err := json.Marshal(data.Content); err == nil {
			msg.Content = contentBytes
		}
	}

	return msg, nil
}
