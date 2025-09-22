// Package constant defines error messages used throughout the chat service.
package constant

const (
	// FailedToPingDatabase indicates an error when the database connection cannot be established.
	FailedToPingDatabase = "failed to ping database: %w"
	// InvalidRequestBodyErrorMessage is the message returned when request body is invalid.
	InvalidRequestBodyErrorMessage = "invalid request body"
	// ConversationNotFoundErrorMessage is the message returned when a conversation is not found.
	ConversationNotFoundErrorMessage = "conversation not found"
	// MessageNotFoundErrorMessage is the message returned when a message is not found.
	MessageNotFoundErrorMessage = "message not found"
	// ParticipantNotFoundErrorMessage is the message returned when a participant is not found.
	ParticipantNotFoundErrorMessage = "participant not found"
	// InvalidConversationIDErrorMessage is the message returned when conversation ID is invalid.
	InvalidConversationIDErrorMessage = "invalid conversation ID"
	// InvalidMessageContentErrorMessage is the message returned when message content is invalid.
	InvalidMessageContentErrorMessage = "message content is required"
	// UserAlreadyParticipantErrorMessage is the message returned when user is already a participant.
	UserAlreadyParticipantErrorMessage = "user is already a participant"
	// UserNotParticipantErrorMessage is the message returned when user is not a participant.
	UserNotParticipantErrorMessage = "user is not a participant"
	// ConversationEndedErrorMessage is the message returned when conversation is already ended.
	ConversationEndedErrorMessage = "conversation has already ended"
	// AccessDeniedErrorMessage is the message returned when access is denied.
	AccessDeniedErrorMessage = "access denied"
)
