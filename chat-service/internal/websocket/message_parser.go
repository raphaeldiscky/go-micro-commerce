package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"
)

// Parser errors.
var (
	ErrInvalidPayload = errors.New("invalid payload")
	ErrMissingID      = errors.New("missing message ID")
	ErrInvalidUUID    = errors.New("invalid UUID format")
	ErrInvalidContent = errors.New("invalid message content")
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
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Channel   *string         `json:"channel,omitempty"`
	SenderID  *string         `json:"sender_id,omitempty"`
	Content   json.RawMessage `json:"content"`
	Timestamp time.Time       `json:"timestamp"`
}

// ParseWebSocketMessage parses a WebSocket message from cross-instance payload.
func (mp *messageParser) ParseWebSocketMessage(
	payload json.RawMessage,
) (*pkgwebsocket.Message, error) {
	if len(payload) == 0 {
		return nil, ErrInvalidPayload
	}

	var data messageData
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPayload, err)
	}

	// Validate required fields
	if data.ID == "" {
		return nil, ErrMissingID
	}

	if data.Type == "" {
		return nil, fmt.Errorf("%w: missing message type", ErrInvalidPayload)
	}

	// Parse and validate ID
	parsedID, err := uuid.Parse(data.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ID %s", ErrInvalidUUID, data.ID)
	}

	msg := &pkgwebsocket.Message{
		ID:        parsedID,
		Type:      pkgwebsocket.MessageType(data.Type),
		Channel:   data.Channel,
		Content:   data.Content,
		Timestamp: data.Timestamp,
	}

	// Parse SenderID if present
	if data.SenderID != nil && *data.SenderID != "" {
		parsedSenderID, senderErr := uuid.Parse(*data.SenderID)
		if senderErr != nil {
			return nil, fmt.Errorf("%w: invalid sender ID %s", ErrInvalidUUID, *data.SenderID)
		}

		msg.SenderID = &parsedSenderID
	}

	return msg, nil
}
