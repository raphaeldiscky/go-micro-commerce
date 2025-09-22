package websocket

import (
	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgwebsocket "github.com/raphaeldiscky/go-micro-commerce/pkg/websocket"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/repository"
)

// ChatHub extends the universal BaseHub with chat-specific functionality.
type ChatHub struct {
	*pkgwebsocket.BaseHub

	logger         logger.Logger
	ConnectionRepo repository.ConnectionRepository
}

// NewChatHub creates a new chat-specific WebSocket hub.
func NewChatHub(connectionRepo repository.ConnectionRepository, logger logger.Logger) *ChatHub {
	baseHub := pkgwebsocket.NewBaseHub(logger)

	return &ChatHub{
		BaseHub:        baseHub,
		ConnectionRepo: connectionRepo,
		logger:         logger,
	}
}

// BroadcastToConversation broadcasts a message to all participants in a conversation.
func (h *ChatHub) BroadcastToConversation(
	conversationID uuid.UUID,
	message *pkgwebsocket.Message,
) error {
	channelName := "conversation:" + conversationID.String()
	return h.BroadcastToChannel(channelName, message)
}

// BroadcastToUserType broadcasts a message to all users of a specific type.
func (h *ChatHub) BroadcastToUserType(
	userType constant.UserType,
	message *pkgwebsocket.Message,
) error {
	filter := func(conn pkgwebsocket.Connection) bool {
		if chatConn, ok := conn.(*ChatConnection); ok {
			return chatConn.UserType() == userType
		}

		return false
	}

	return h.Broadcast(message, filter)
}

// JoinConversation adds a connection to a conversation channel.
func (h *ChatHub) JoinConversation(conn *ChatConnection, conversationID uuid.UUID) {
	channelName := "conversation:" + conversationID.String()
	h.JoinChannel(conn, channelName)
	conn.JoinConversation(conversationID)
}

// LeaveConversation removes a connection from a conversation channel.
func (h *ChatHub) LeaveConversation(conn *ChatConnection, conversationID uuid.UUID) {
	channelName := "conversation:" + conversationID.String()
	h.LeaveChannel(conn, channelName)
	conn.LeaveConversation()
}

// GetConversationConnections returns all connections in a conversation.
func (h *ChatHub) GetConversationConnections(conversationID uuid.UUID) []*ChatConnection {
	channelName := "conversation:" + conversationID.String()
	universalConns := h.GetChannelConnections(channelName)

	chatConns := make([]*ChatConnection, 0, len(universalConns))
	for _, conn := range universalConns {
		if chatConn, ok := conn.(*ChatConnection); ok {
			chatConns = append(chatConns, chatConn)
		}
	}

	return chatConns
}

// GetUserTypeConnections returns all connections for a specific user type.
func (h *ChatHub) GetUserTypeConnections(userType constant.UserType) []*ChatConnection {
	connections := make([]*ChatConnection, 0)

	// Get all user connections and filter by type
	err := h.BaseHub.Broadcast(&pkgwebsocket.Message{}, func(conn pkgwebsocket.Connection) bool {
		if chatConn, ok := conn.(*ChatConnection); ok && chatConn.UserType() == userType {
			connections = append(connections, chatConn)
		}

		return false // Don't actually send the message, just collect connections
	})
	if err != nil {
		h.logger.Error("Failed to broadcast message", "error", err)
	}

	return connections
}

// GetConnectionCount returns the total number of active connections.
func (h *ChatHub) GetConnectionCount() int {
	return h.BaseHub.GetConnectionCount()
}

// GetUserCount returns the number of unique users connected.
func (h *ChatHub) GetUserCount() int {
	return h.BaseHub.GetUserCount()
}
