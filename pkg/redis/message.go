package redis

import (
	"fmt"

	"github.com/bytedance/sonic"
)

// Message represents a pub/sub message with metadata and payload.
type Message struct {
	// Metadata contains message metadata.
	Metadata MessageMetadata `json:"metadata"`
	// Payload contains the actual message data.
	Payload sonic.NoCopyRawMessage `json:"payload"`
}

// NewMessage creates a new message with the given metadata and payload.
func NewMessage(metadata MessageMetadata, payload any) (*Message, error) {
	payloadBytes, err := sonic.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return &Message{
		Metadata: metadata,
		Payload:  payloadBytes,
	}, nil
}

// UnmarshalPayload unmarshals the message payload into the provided target.
func (m *Message) UnmarshalPayload(target any) error {
	if err := sonic.Unmarshal(m.Payload, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}

// ToJSON serializes the message to JSON bytes.
func (m *Message) ToJSON() ([]byte, error) {
	data, err := sonic.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return data, nil
}

// FromJSON deserializes a message from JSON bytes.
func FromJSON(data []byte) (*Message, error) {
	var message Message
	if err := sonic.Unmarshal(data, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &message, nil
}

// GetCorrelationID returns the correlation ID from message metadata.
func (m *Message) GetCorrelationID() string {
	return m.Metadata.CorrelationID
}

// GetSource returns the source service from message metadata.
func (m *Message) GetSource() string {
	return m.Metadata.Source
}

// GetMessageID returns the message ID from metadata.
func (m *Message) GetMessageID() string {
	return m.Metadata.MessageID
}
