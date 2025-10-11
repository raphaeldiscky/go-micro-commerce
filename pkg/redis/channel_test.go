package redis_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/redis"
)

func TestChannelBuilder(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *redis.ChannelBuilder
		expected string
	}{
		{
			name: "simple service channel",
			builder: func() *redis.ChannelBuilder {
				return redis.NewChannelBuilder().Service("chat")
			},
			expected: "chat",
		},
		{
			name: "service with domain",
			builder: func() *redis.ChannelBuilder {
				return redis.NewChannelBuilder().Service("chat").Domain("messaging")
			},
			expected: "chat:messaging",
		},
		{
			name: "full entity channel",
			builder: func() *redis.ChannelBuilder {
				return redis.NewChannelBuilder().
					Service("chat").
					Entity("conversation").
					IDString("123")
			},
			expected: "chat:conversation:123",
		},
		{
			name: "channel with event",
			builder: func() *redis.ChannelBuilder {
				return redis.NewChannelBuilder().
					Service("order").
					Entity("order").
					IDString("order-456").
					Event("created")
			},
			expected: "order:order:order-456:created",
		},
		{
			name: "channel with custom parts",
			builder: func() *redis.ChannelBuilder {
				return redis.NewChannelBuilder().
					Service("notification").
					Custom("broadcast").
					Custom("urgent")
			},
			expected: "notification:broadcast:urgent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.builder().Build()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChannelBuilder_Pattern(t *testing.T) {
	pattern := redis.NewChannelBuilder().
		Service("chat").
		Entity("conversation").
		BuildPattern()

	assert.Equal(t, "chat:conversation*", pattern)
}

func TestChatChannel(t *testing.T) {
	conversationID := uuid.New()
	channel := redis.ChatChannel(conversationID)

	expected := "chat:conversation:" + conversationID.String()
	assert.Equal(t, expected, channel)
}

func TestUserPresenceChannel(t *testing.T) {
	userID := uuid.New()
	channel := redis.UserPresenceChannel(userID)

	expected := "chat:presence:" + userID.String()
	assert.Equal(t, expected, channel)
}

func TestBroadcastChannel(t *testing.T) {
	messageType := "maintenance"
	channel := redis.BroadcastChannel(messageType)

	expected := "system:broadcast:maintenance"
	assert.Equal(t, expected, channel)
}

func TestOrderChannel(t *testing.T) {
	orderID := uuid.New()
	eventType := "status_changed"
	channel := redis.OrderChannel(orderID, eventType)

	expected := "order:order:" + orderID.String() + ":status_changed"
	assert.Equal(t, expected, channel)
}

func TestAdminChannel(t *testing.T) {
	adminID := uuid.New()
	channelType := "alerts"
	channel := redis.AdminChannel(adminID, channelType)

	expected := "admin:admin:" + adminID.String() + ":alerts"
	assert.Equal(t, expected, channel)
}

func TestServiceHealthChannel(t *testing.T) {
	serviceName := "payment-service"
	channel := redis.ServiceHealthChannel(serviceName)

	expected := "monitoring:health:payment-service"
	assert.Equal(t, expected, channel)
}

func TestPatternBuilders(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		expected string
	}{
		{
			name:     "all chat pattern",
			function: redis.AllChatPattern,
			expected: "chat*",
		},
		{
			name:     "all notification pattern",
			function: redis.AllNotificationPattern,
			expected: "notification*",
		},
		{
			name:     "all order pattern",
			function: redis.AllOrderPattern,
			expected: "order*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserChatPattern(t *testing.T) {
	userID := uuid.New()
	pattern := redis.UserChatPattern(userID)

	expected := "chat:conversation:*:" + userID.String() + ":*"
	assert.Equal(t, expected, pattern)
}
