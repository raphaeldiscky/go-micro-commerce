package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/repository"
)

// ConversationAccess provides conversation access verification functionality.
type ConversationAccess interface {
	// VerifyUserAccess verifies if a user has access to a conversation.
	VerifyUserAccess(ctx context.Context, conversationID, userID uuid.UUID) error
	// VerifyActiveUserAccess verifies if a user has active access to a conversation.
	VerifyActiveUserAccess(ctx context.Context, conversationID, userID uuid.UUID) error
}

type conversationAccess struct {
	dataStore repository.DataStore
}

// NewConversationAccess creates a new conversation access service.
func NewConversationAccess(dataStore repository.DataStore) ConversationAccess {
	return &conversationAccess{
		dataStore: dataStore,
	}
}

// VerifyUserAccess verifies if a user has access to a conversation (active or inactive).
func (ca *conversationAccess) VerifyUserAccess(
	ctx context.Context,
	conversationID, userID uuid.UUID,
) error {
	participantRepo := ca.dataStore.ParticipantRepository()

	participants, err := participantRepo.FindByConversationID(ctx, conversationID)
	if err != nil {
		return httperror.NewInternalServerError("failed to check conversation access")
	}

	for _, p := range participants {
		if p.UserID == userID {
			return nil
		}
	}

	return httperror.NewBadRequestError("access denied to conversation")
}

// VerifyActiveUserAccess verifies if a user has active access to a conversation.
func (ca *conversationAccess) VerifyActiveUserAccess(
	ctx context.Context,
	conversationID, userID uuid.UUID,
) error {
	participantRepo := ca.dataStore.ParticipantRepository()

	participants, err := participantRepo.FindByConversationID(ctx, conversationID)
	if err != nil {
		return httperror.NewInternalServerError("failed to check conversation access")
	}

	for _, p := range participants {
		if p.UserID == userID && p.IsActive {
			return nil
		}
	}

	return httperror.NewBadRequestError("access denied to conversation")
}
