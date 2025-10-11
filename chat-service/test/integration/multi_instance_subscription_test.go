package integration_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

const (
	messageReceiveTimeout = 5 * time.Second
	subscriptionWaitTime  = 1 * time.Second
)

// MultiInstanceTestSuite holds the test suite for multi-instance scenarios.
type MultiInstanceTestSuite struct {
	TestSuite
}

// TestDynamicSubscription tests that instances only subscribe to channels with active connections.
func (s *MultiInstanceTestSuite) TestDynamicSubscription() {
	// Start 3 chat-service instances
	instance1, err := s.StartChatServiceInstance(8090)
	s.Require().NoError(err)

	defer func() {
		_ = s.StopChatServiceInstance(instance1)
	}()

	instance2, err := s.StartChatServiceInstance(8091)
	s.Require().NoError(err)

	defer func() {
		_ = s.StopChatServiceInstance(instance2)
	}()

	instance3, err := s.StartChatServiceInstance(8092)
	s.Require().NoError(err)

	defer func() {
		_ = s.StopChatServiceInstance(instance3)
	}()

	// Create test users
	user1 := uuid.New()
	user2 := uuid.New()

	// Create test conversation
	_, err = s.CreateTestConversation(instance1, []uuid.UUID{user1, user2})
	s.Require().NoError(err)

	// Verify no active subscriptions initially
	s.Equal(0, instance1.Hub.GetActiveChannelCount(), "Instance 1 should have no subscriptions")
	s.Equal(0, instance2.Hub.GetActiveChannelCount(), "Instance 2 should have no subscriptions")
	s.Equal(0, instance3.Hub.GetActiveChannelCount(), "Instance 3 should have no subscriptions")

	s.logger.Info("✓ Initial state: No active subscriptions")

	// Connect user1 to instance1 (auto-join happens on connect)
	conn1, err := s.ConnectWebSocket(instance1, user1, constant.UserTypeUser)
	s.Require().NoError(err)

	defer conn1.Close()

	// Wait for auto-join to complete
	time.Sleep(subscriptionWaitTime)

	// Verify only instance1 has active subscription
	s.Equal(1, instance1.Hub.GetActiveChannelCount(), "Instance 1 should have 1 subscription")
	s.Equal(0, instance2.Hub.GetActiveChannelCount(), "Instance 2 should have no subscriptions")
	s.Equal(0, instance3.Hub.GetActiveChannelCount(), "Instance 3 should have no subscriptions")

	s.logger.Info("✓ Dynamic subscription: Instance 1 subscribed on first connection")

	// Connect user2 to instance2 (different instance, auto-join happens)
	conn2, err := s.ConnectWebSocket(instance2, user2, constant.UserTypeUser)
	s.Require().NoError(err)

	defer conn2.Close()

	// Wait for auto-join
	time.Sleep(subscriptionWaitTime)

	// Verify both instance1 and instance2 have subscriptions, but not instance3
	s.Equal(1, instance1.Hub.GetActiveChannelCount(), "Instance 1 should have 1 subscription")
	s.Equal(1, instance2.Hub.GetActiveChannelCount(), "Instance 2 should have 1 subscription")
	s.Equal(0, instance3.Hub.GetActiveChannelCount(), "Instance 3 should have no subscriptions")

	s.logger.Info("✓ Dynamic subscription: Instance 2 subscribed on first connection")

	// Disconnect user1 from instance1
	err = conn1.Close()
	s.Require().NoError(err)

	// Wait for unsubscription
	time.Sleep(subscriptionWaitTime)

	// Verify instance1 unsubscribed, but instance2 still subscribed
	s.Equal(
		0,
		instance1.Hub.GetActiveChannelCount(),
		"Instance 1 should have no subscriptions after last connection left",
	)
	s.Equal(1, instance2.Hub.GetActiveChannelCount(), "Instance 2 should still have 1 subscription")

	s.logger.Info("✓ Dynamic unsubscription: Instance 1 unsubscribed when last connection left")
}

// TestCrossInstanceMessageDelivery tests that messages are delivered across instances.
func (s *MultiInstanceTestSuite) TestCrossInstanceMessageDelivery() {
	// Start 2 instances
	instance1, err := s.StartChatServiceInstance(8090)
	s.Require().NoError(err)

	defer func() {
		_ = s.StopChatServiceInstance(instance1)
	}()

	instance2, err := s.StartChatServiceInstance(8091)
	s.Require().NoError(err)

	defer func() {
		_ = s.StopChatServiceInstance(instance2)
	}()

	// Create test users
	user1 := uuid.New()
	user2 := uuid.New()

	// Create conversation
	conversation, err := s.CreateTestConversation(instance1, []uuid.UUID{user1, user2})
	s.Require().NoError(err)

	// Connect user1 to instance1 (auto-join happens)
	conn1, err := s.ConnectWebSocket(instance1, user1, constant.UserTypeUser)
	s.Require().NoError(err)

	defer conn1.Close()

	// Connect user2 to instance2 (different instance, auto-join happens)
	conn2, err := s.ConnectWebSocket(instance2, user2, constant.UserTypeUser)
	s.Require().NoError(err)

	defer conn2.Close()

	// Wait for auto-join subscriptions
	time.Sleep(subscriptionWaitTime)

	s.logger.Info("Both users joined conversation on different instances",
		"instance1_subscriptions", instance1.Hub.GetActiveChannelCount(),
		"instance2_subscriptions", instance2.Hub.GetActiveChannelCount())

	// Send message from user1 (instance1)
	chatContent := websocket.ChatContent{
		ConversationID: conversation.ID,
		Text:           "Hello from instance 1!",
		MessageType:    constant.MessageTypeText,
	}

	err = s.SendWebSocketMessage(conn1, websocket.ChatMessageTypeChat, chatContent)
	s.Require().NoError(err)

	s.logger.Info("Message sent from user1 on instance1")

	// User2 should receive the message on instance2
	msg, err := s.ReceiveWebSocketMessage(conn2, messageReceiveTimeout)
	s.Require().NoError(err, "User2 should receive message from user1")
	s.Equal(websocket.ChatMessageTypeChat, msg.Type)

	// Verify message content
	var receivedContent websocket.ChatContent

	err = json.Unmarshal(msg.Content, &receivedContent)
	s.Require().NoError(err)
	s.Equal("Hello from instance 1!", receivedContent.Text)

	s.logger.Info("✓ Cross-instance delivery: Message delivered from instance1 to instance2")

	// Send reply from user2 (instance2)
	replyContent := websocket.ChatContent{
		ConversationID: conversation.ID,
		Text:           "Hello from instance 2!",
		MessageType:    constant.MessageTypeText,
	}

	err = s.SendWebSocketMessage(conn2, websocket.ChatMessageTypeChat, replyContent)
	s.Require().NoError(err)

	// User1 should receive the reply on instance1
	msg, err = s.ReceiveWebSocketMessage(conn1, messageReceiveTimeout)
	s.Require().NoError(err, "User1 should receive message from user2")
	s.Equal(websocket.ChatMessageTypeChat, msg.Type)

	// Verify reply content
	err = json.Unmarshal(msg.Content, &receivedContent)
	s.Require().NoError(err)
	s.Equal("Hello from instance 2!", receivedContent.Text)

	s.logger.Info("✓ Cross-instance delivery: Reply delivered from instance2 to instance1")
}

// TestNoMessageLoops verifies that instance ID filtering prevents message loops.
func (s *MultiInstanceTestSuite) TestNoMessageLoops() {
	// Start 1 instance
	instance1, err := s.StartChatServiceInstance(8090)
	s.Require().NoError(err)

	defer func() {
		_ = s.StopChatServiceInstance(instance1)
	}()

	// Create test users
	user1 := uuid.New()
	user2 := uuid.New()

	// Create conversation
	conversation, err := s.CreateTestConversation(instance1, []uuid.UUID{user1, user2})
	s.Require().NoError(err)

	// Connect both users to the same instance (auto-join happens)
	conn1, err := s.ConnectWebSocket(instance1, user1, constant.UserTypeUser)
	s.Require().NoError(err)

	defer conn1.Close()

	conn2, err := s.ConnectWebSocket(instance1, user2, constant.UserTypeUser)
	s.Require().NoError(err)

	defer conn2.Close()

	// Wait for auto-join
	time.Sleep(subscriptionWaitTime)

	// Send message from user1
	chatContent := websocket.ChatContent{
		ConversationID: conversation.ID,
		Text:           "Test message",
		MessageType:    constant.MessageTypeText,
	}

	err = s.SendWebSocketMessage(conn1, websocket.ChatMessageTypeChat, chatContent)
	s.Require().NoError(err)

	// User2 should receive the message
	msg, err := s.ReceiveWebSocketMessage(conn2, messageReceiveTimeout)
	s.Require().NoError(err)
	s.Equal(websocket.ChatMessageTypeChat, msg.Type)

	// User1 should NOT receive their own message back via Redis
	// (they already got it via local broadcast)
	// Set a short timeout and expect timeout error
	_, err = s.ReceiveWebSocketMessage(conn1, 2*time.Second)
	s.Error(err, "User1 should not receive their own message via Redis loop")

	s.logger.Info("✓ No message loops: Instance ID filtering works correctly")
}

// TestMultipleConversationsScaling tests dynamic subscription with many conversations.
func (s *MultiInstanceTestSuite) TestMultipleConversationsScaling() {
	// Start 2 instances
	instance1, err := s.StartChatServiceInstance(8090)
	s.Require().NoError(err)

	defer func() {
		_ = s.StopChatServiceInstance(instance1)
	}()

	instance2, err := s.StartChatServiceInstance(8091)
	s.Require().NoError(err)

	defer func() {
		_ = s.StopChatServiceInstance(instance2)
	}()

	// Create 10 conversations
	const numConversations = 10

	conversations := make([]uuid.UUID, numConversations)

	// Create a single user that will participate in all conversations
	user := uuid.New()

	for i := range numConversations {
		otherUser := uuid.New()

		conv, err := s.CreateTestConversation(instance1, []uuid.UUID{user, otherUser})
		s.Require().NoError(err)

		conversations[i] = conv.ID
	}

	// Verify no subscriptions initially
	s.Equal(0, instance1.Hub.GetActiveChannelCount())
	s.Equal(0, instance2.Hub.GetActiveChannelCount())

	// Connect user to instance1 (auto-join will join all 10 conversations)
	conn1, err := s.ConnectWebSocket(instance1, user, constant.UserTypeUser)
	s.Require().NoError(err)

	defer conn1.Close()

	time.Sleep(subscriptionWaitTime)

	// Verify instance1 has 10 subscriptions (one for each conversation)
	s.Equal(numConversations, instance1.Hub.GetActiveChannelCount(),
		"Instance 1 should have 10 active subscriptions")

	s.logger.Info("✓ Scaling: Instance 1 has correct number of subscriptions",
		"subscriptions", instance1.Hub.GetActiveChannelCount())

	// Connect same user to instance2 (auto-join will join all 10 conversations there too)
	conn2, err := s.ConnectWebSocket(instance2, user, constant.UserTypeUser)
	s.Require().NoError(err)

	defer conn2.Close()

	time.Sleep(subscriptionWaitTime)

	// Verify instance2 also has 10 subscriptions
	s.Equal(numConversations, instance2.Hub.GetActiveChannelCount(),
		"Instance 2 should have 10 active subscriptions")

	s.logger.Info("✓ Scaling: Instance 2 has correct number of subscriptions",
		"subscriptions", instance2.Hub.GetActiveChannelCount())

	// Total subscriptions across all instances = 20 (10 per instance)
	totalSubscriptions := instance1.Hub.GetActiveChannelCount() + instance2.Hub.GetActiveChannelCount()
	s.Equal(numConversations*2, totalSubscriptions,
		"Total subscriptions should equal number of conversations times instances")

	s.logger.Info("✓ Scaling test passed",
		"total_conversations", numConversations,
		"total_subscriptions", totalSubscriptions,
		"instance1", instance1.Hub.GetActiveChannelCount(),
		"instance2", instance2.Hub.GetActiveChannelCount())
}

// Entrypoint to run the multi-instance test suite.
func TestMultiInstanceSuite(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("Skipping multi-instance integration tests in short mode")
	}

	fmt.Println("🚀 Starting multi-instance dynamic subscription tests...")
	suite.Run(t, new(MultiInstanceTestSuite))
}
