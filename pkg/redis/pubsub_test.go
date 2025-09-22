package redis_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
)

func TestNewMessageMetadata(t *testing.T) {
	source := "test-service"
	metadata := redis.NewMessageMetadata(source)

	assert.Equal(t, source, metadata.Source)
	assert.NotEmpty(t, metadata.MessageID)
	assert.Equal(t, "application/json", metadata.ContentType)
	assert.WithinDuration(t, time.Now(), metadata.Timestamp, time.Second)
}

func TestMessageMetadata_SetCorrelationID(t *testing.T) {
	metadata := redis.NewMessageMetadata("test-service")
	correlationID := uuid.New().String()

	metadata.SetCorrelationID(correlationID)

	assert.Equal(t, correlationID, metadata.CorrelationID)
}

func TestDefaultPubSubConfig(t *testing.T) {
	config := redis.DefaultPubSubConfig()

	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, 100*time.Millisecond, config.RetryDelay)
	assert.Equal(t, 5*time.Second, config.MaxRetryDelay)
	assert.Equal(t, 100, config.ChannelBufferSize)
}
