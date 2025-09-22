package websocket

import (
	"fmt"

	"github.com/google/uuid"
)

// ConversationChannel generates a channel name for a conversation.
func ConversationChannel(conversationID uuid.UUID) string {
	return fmt.Sprintf("conversation:%s", conversationID.String())
}

// UserChannel generates a channel name for a user.
func UserChannel(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s", userID.String())
}
