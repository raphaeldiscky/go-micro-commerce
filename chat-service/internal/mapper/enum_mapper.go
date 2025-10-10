package mapper

import (
	"strings"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
)

// NormalizeMessageType converts GraphQL enum (UPPERCASE) to database format (lowercase).
// GraphQL: TEXT, IMAGE, FILE, SYSTEM
// Database: text, image, file, system.
func NormalizeMessageType(graphqlEnum constant.MessageType) constant.MessageType {
	return constant.MessageType(strings.ToLower(string(graphqlEnum)))
}

// NormalizeConversationStatus converts GraphQL enum (UPPERCASE) to database format (lowercase).
// GraphQL: WAITING, ACTIVE, ENDED
// Database: waiting, active, ended.
func NormalizeConversationStatus(
	graphqlEnum constant.ConversationStatus,
) constant.ConversationStatus {
	return constant.ConversationStatus(strings.ToLower(string(graphqlEnum)))
}

// NormalizeParticipantRole converts GraphQL enum (UPPERCASE) to database format (lowercase).
// GraphQL: OWNER, MODERATOR, MEMBER
// Database: owner, moderator, member.
func NormalizeParticipantRole(graphqlEnum constant.ParticipantRole) constant.ParticipantRole {
	return constant.ParticipantRole(strings.ToLower(string(graphqlEnum)))
}

// NormalizePresenceStatus converts GraphQL enum (UPPERCASE) to database format (lowercase).
// GraphQL: ONLINE, OFFLINE, AWAY, BUSY
// Database: online, offline, away, busy.
func NormalizePresenceStatus(graphqlEnum constant.PresenceStatus) constant.PresenceStatus {
	return constant.PresenceStatus(strings.ToLower(string(graphqlEnum)))
}
