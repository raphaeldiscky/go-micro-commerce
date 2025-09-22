package constant

// ConversationStatus represents the status of a chat conversation.
type ConversationStatus string

const (
	// ConversationStatusWaiting represents a conversation waiting for admin assignment.
	ConversationStatusWaiting ConversationStatus = "waiting"
	// ConversationStatusActive represents an active conversation.
	ConversationStatusActive ConversationStatus = "active"
	// ConversationStatusEnded represents an ended conversation.
	ConversationStatusEnded ConversationStatus = "ended"
	// ConversationStatusTransferred represents a conversation transferred to another admin.
	ConversationStatusTransferred ConversationStatus = "transferred"
)

// UserType represents the type of user in the chat system.
type UserType string

const (
	// UserTypeUser represents a regular user/customer.
	UserTypeUser UserType = "user"
	// UserTypeAdmin represents an admin/support agent.
	UserTypeAdmin UserType = "admin"
)

// MessageType represents the type of message.
type MessageType string

const (
	// MessageTypeText represents a text message.
	MessageTypeText MessageType = "text"
	// MessageTypeImage represents an image message.
	MessageTypeImage MessageType = "image"
	// MessageTypeFile represents a file attachment message.
	MessageTypeFile MessageType = "file"
	// MessageTypeSystem represents a system-generated message.
	MessageTypeSystem MessageType = "system"
)

// ParticipantRole represents the role of a participant in a conversation.
type ParticipantRole string

const (
	// ParticipantRoleParticipant represents a regular participant.
	ParticipantRoleParticipant ParticipantRole = "participant"
	// ParticipantRoleModerator represents a moderator with elevated permissions.
	ParticipantRoleModerator ParticipantRole = "moderator"
	// ParticipantRoleObserver represents an observer with read-only access.
	ParticipantRoleObserver ParticipantRole = "observer"
)

// Priority levels for conversations.
const (
	// PriorityLow represents low priority conversations.
	PriorityLow = 1
	// PriorityNormal represents normal priority conversations.
	PriorityNormal = 2
	// PriorityHigh represents high priority conversations.
	PriorityHigh = 3
	// PriorityUrgent represents urgent priority conversations.
	PriorityUrgent = 4
)

// PresenceStatus represents user presence status.
type PresenceStatus string

const (
	// PresenceStatusOnline represents an online user.
	PresenceStatusOnline PresenceStatus = "online"
	// PresenceStatusOffline represents an offline user.
	PresenceStatusOffline PresenceStatus = "offline"
	// PresenceStatusAway represents an away user.
	PresenceStatusAway PresenceStatus = "away"
	// PresenceStatusBusy represents a busy user.
	PresenceStatusBusy PresenceStatus = "busy"
)

// WebSocketEventType represents the type of WebSocket event.
type WebSocketEventType string

const (
	// WebSocketEventTypeConnect represents a connection event.
	WebSocketEventTypeConnect WebSocketEventType = "connect"
	// WebSocketEventTypeDisconnect represents a disconnection event.
	WebSocketEventTypeDisconnect WebSocketEventType = "disconnect"
	// WebSocketEventTypeJoin represents joining a conversation.
	WebSocketEventTypeJoin WebSocketEventType = "join"
	// WebSocketEventTypeLeave represents leaving a conversation.
	WebSocketEventTypeLeave WebSocketEventType = "leave"
)

// WebSocket configuration constants.
const (
	// WebSocketReadBufferSize defines the read buffer size for WebSocket connections.
	WebSocketReadBufferSize = 1024
	// WebSocketWriteBufferSize defines the write buffer size for WebSocket connections.
	WebSocketWriteBufferSize = 1024
	// WebSocketMaxMessageSize defines the maximum message size allowed.
	WebSocketMaxMessageSize = 512
	// WebSocketPongWait defines the timeout for receiving pong messages.
	WebSocketPongWait = 60
	// WebSocketPingPeriod defines the interval for sending ping messages.
	WebSocketPingPeriod = 54
	// WebSocketWriteWait defines the timeout for writing messages.
	WebSocketWriteWait = 10
	// WebSocketSendBufferSize defines the channel buffer size for sending messages.
	WebSocketSendBufferSize = 256
)

// Service configuration constants.
const (
	// DefaultShutdownTimeout defines the default timeout for service shutdown.
	DefaultShutdownTimeout = 30
	// DefaultMessageLimit defines the default limit for message queries.
	DefaultMessageLimit = 50
	// DefaultConversationLimit defines the default limit for conversation queries.
	DefaultConversationLimit = 10000
)
