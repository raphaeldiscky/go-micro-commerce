package redis_test

import (
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
)

type testPayload struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

func TestNewMessage(t *testing.T) {
	metadata := redis.NewMessageMetadata("test-service")
	payload := testPayload{
		ID:      "test-123",
		Message: "Hello, World!",
		Count:   42,
	}

	message, err := redis.NewMessage(metadata, payload)

	require.NoError(t, err)
	assert.Equal(t, metadata, message.Metadata)
	assert.NotNil(t, message.Payload)

	// Verify payload can be unmarshaled
	var unmarshaled testPayload

	err = message.UnmarshalPayload(&unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, payload, unmarshaled)
}

func TestMessage_UnmarshalPayload(t *testing.T) {
	originalPayload := testPayload{
		ID:      "test-456",
		Message: "Test message",
		Count:   100,
	}

	metadata := redis.NewMessageMetadata("test-service")
	message, err := redis.NewMessage(metadata, originalPayload)
	require.NoError(t, err)

	var unmarshaled testPayload

	err = message.UnmarshalPayload(&unmarshaled)

	require.NoError(t, err)
	assert.Equal(t, originalPayload, unmarshaled)
}

func TestMessage_ToJSON(t *testing.T) {
	metadata := redis.NewMessageMetadata("test-service")
	payload := testPayload{
		ID:      "test-789",
		Message: "JSON test",
		Count:   25,
	}

	message, err := redis.NewMessage(metadata, payload)
	require.NoError(t, err)

	jsonData, err := message.ToJSON()

	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify it's valid JSON
	var jsonMap map[string]any

	err = sonic.Unmarshal(jsonData, &jsonMap)
	require.NoError(t, err)
	assert.Contains(t, jsonMap, "metadata")
	assert.Contains(t, jsonMap, "payload")
}

func TestFromJSON(t *testing.T) {
	originalPayload := testPayload{
		ID:      "test-from-json",
		Message: "Round trip test",
		Count:   75,
	}

	metadata := redis.NewMessageMetadata("test-service")
	originalMessage, err := redis.NewMessage(metadata, originalPayload)
	require.NoError(t, err)

	jsonData, err := originalMessage.ToJSON()
	require.NoError(t, err)

	restoredMessage, err := redis.FromJSON(jsonData)

	require.NoError(t, err)
	assert.Equal(t, originalMessage.Metadata.MessageID, restoredMessage.Metadata.MessageID)
	assert.Equal(t, originalMessage.Metadata.Source, restoredMessage.Metadata.Source)
	assert.Equal(t, originalMessage.Metadata.ContentType, restoredMessage.Metadata.ContentType)
	assert.WithinDuration(
		t,
		originalMessage.Metadata.Timestamp,
		restoredMessage.Metadata.Timestamp,
		time.Second,
	)

	var restoredPayload testPayload

	err = restoredMessage.UnmarshalPayload(&restoredPayload)
	require.NoError(t, err)
	assert.Equal(t, originalPayload, restoredPayload)
}

func TestMessage_GetMethods(t *testing.T) {
	metadata := redis.NewMessageMetadata("test-service")
	metadata.SetCorrelationID("correlation-123")

	payload := testPayload{ID: "test"}
	message, err := redis.NewMessage(metadata, payload)
	require.NoError(t, err)

	assert.Equal(t, metadata.CorrelationID, message.GetCorrelationID())
	assert.Equal(t, metadata.Source, message.GetSource())
	assert.Equal(t, metadata.MessageID, message.GetMessageID())
}

func TestFromJSON_InvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{"invalid": json`)

	message, err := redis.FromJSON(invalidJSON)

	require.Error(t, err)
	assert.Nil(t, message)
	assert.Contains(t, err.Error(), "failed to unmarshal message")
}
