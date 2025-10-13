// Package service provides business logic for chat operations.
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/httperror"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/mapper"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/repository"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/websocket"
)

// ChatService defines the interface for chat business operations.
type ChatService interface {
	// Conversation management
	CreateConversation(
		ctx context.Context,
		userID uuid.UUID,
		userType constant.UserType,
		req *dto.CreateConversationRequest,
	) (*dto.ConversationResponse, error)
	GetConversation(
		ctx context.Context,
		conversationID uuid.UUID,
		userID uuid.UUID,
	) (*dto.ConversationResponse, error)
	GetConversationByID(
		ctx context.Context,
		conversationID uuid.UUID,
	) (*dto.ConversationResponse, error)
	GetUserConversations(
		ctx context.Context,
		userID uuid.UUID,
		userType constant.UserType,
	) ([]dto.ConversationResponse, error)
	GetUserConversationsWithCursor(
		ctx context.Context,
		userID uuid.UUID,
		userType constant.UserType,
		limit int,
		afterCursor string,
		beforeCursor string,
	) ([]dto.ConversationResponse, *pkgdto.CursorPagination, error)
	EndConversation(
		ctx context.Context,
		conversationID uuid.UUID,
	) (*dto.ConversationResponse, error)

	// Message management (read-only - sending handled via WebSocket)
	GetConversationMessages(
		ctx context.Context,
		conversationID uuid.UUID,
		userID uuid.UUID,
		limit, offset int,
	) ([]dto.MessageResponse, *pkgdto.OffsetPagination, error)
	GetConversationMessagesWithCursor(
		ctx context.Context,
		conversationID uuid.UUID,
		userID uuid.UUID,
		limit int,
		afterCursor string,
		beforeCursor string,
	) ([]dto.MessageResponse, *pkgdto.CursorPagination, error)
	GetMessageByID(ctx context.Context, messageID uuid.UUID) (*dto.MessageResponse, error)

	// Participant management
	JoinConversation(
		ctx context.Context,
		conversationID uuid.UUID,
		userID uuid.UUID,
		userType constant.UserType,
		role constant.ParticipantRole,
	) (*dto.ParticipantResponse, error)
	LeaveConversation(ctx context.Context, conversationID uuid.UUID, userID uuid.UUID) error
	GetConversationParticipants(
		ctx context.Context,
		conversationID uuid.UUID,
	) ([]dto.ParticipantResponse, error)

	// Admin operations
	AssignConversationToAdmin(
		ctx context.Context,
		conversationID uuid.UUID,
		adminID uuid.UUID,
	) (*dto.ConversationResponse, error)
	GetWaitingConversations(ctx context.Context) ([]dto.ConversationResponse, error)
}

// chatService implements the ChatService interface.
type chatService struct {
	dataStore          repository.DataStore
	logger             logger.Logger
	hub                *websocket.ChatHub
	conversationAccess ConversationAccess
}

// NewChatService creates a new instance of chatService.
func NewChatService(
	dataStore repository.DataStore,
	appLogger logger.Logger,
	hub *websocket.ChatHub,
) ChatService {
	conversationAccess := NewConversationAccess(dataStore)

	return &chatService{
		dataStore:          dataStore,
		logger:             appLogger,
		hub:                hub,
		conversationAccess: conversationAccess,
	}
}

// CreateConversation creates a new conversation.
func (s *chatService) CreateConversation(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
	req *dto.CreateConversationRequest,
) (*dto.ConversationResponse, error) {
	var result *dto.ConversationResponse

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		conversationRepo := ds.ConversationRepository()
		participantRepo := ds.ParticipantRepository()

		// Create conversation entity
		conversation, err := entity.NewConversation(req.Subject, req.Priority)
		if err != nil {
			return httperror.NewBadRequestError(
				fmt.Sprintf("failed to create conversation: %v", err),
			)
		}

		s.logger.Infof("Creating conversation for user %s: %v", userID, conversation)
		// Save conversation
		savedConversation, err := conversationRepo.Create(ctx, conversation)
		if err != nil {
			s.logger.Errorf("Failed to save conversation: %v", err)
			return httperror.NewInternalServerError("failed to save conversation")
		}

		// Add creator as participant
		participant, err := entity.NewParticipant(
			savedConversation.ID,
			userID,
			userType,
			constant.ParticipantRoleParticipant,
		)
		if err != nil {
			return httperror.NewBadRequestError(
				fmt.Sprintf("failed to create participant: %v", err),
			)
		}

		_, err = participantRepo.Create(ctx, participant)
		if err != nil {
			s.logger.Errorf("Failed to add participant: %v", err)
			return httperror.NewInternalServerError("failed to add participant")
		}

		result = mapper.MapToConversationResponse(savedConversation)

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.Infof("Created conversation %s for user %s", result.ID, userID)

	return result, nil
}

// GetConversation retrieves a conversation by ID.
func (s *chatService) GetConversation(
	ctx context.Context,
	conversationID uuid.UUID,
	userID uuid.UUID,
) (*dto.ConversationResponse, error) {
	conversationRepo := s.dataStore.ConversationRepository()

	// Check if user is participant
	if err := s.conversationAccess.VerifyActiveUserAccess(ctx, conversationID, userID); err != nil {
		return nil, err
	}

	// Get conversation
	conversation, err := conversationRepo.FindByID(ctx, conversationID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get conversation")
	}

	if conversation == nil {
		return nil, httperror.NewBadRequestError("conversation not found")
	}

	return mapper.MapToConversationResponse(conversation), nil
}

// GetConversationByID retrieves a conversation by ID without access control (for federation).
func (s *chatService) GetConversationByID(
	ctx context.Context,
	conversationID uuid.UUID,
) (*dto.ConversationResponse, error) {
	conversationRepo := s.dataStore.ConversationRepository()

	conversation, err := conversationRepo.FindByID(ctx, conversationID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get conversation")
	}

	if conversation == nil {
		return nil, httperror.NewBadRequestError("conversation not found")
	}

	return mapper.MapToConversationResponse(conversation), nil
}

// GetUserConversations retrieves all conversations for a user.
func (s *chatService) GetUserConversations(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
) ([]dto.ConversationResponse, error) {
	participantRepo := s.dataStore.ParticipantRepository()
	conversationRepo := s.dataStore.ConversationRepository()

	// Get user's active participations with detailed logging
	s.logger.Info("Getting user conversations",
		"user_id", userID,
		"user_type", userType,
		"user_type_string", string(userType))

	participants, err := participantRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find active participants",
			"user_id", userID,
			"user_type", userType,
			"user_type_string", string(userType),
			"error", err)

		return nil, httperror.NewInternalServerError("failed to get user conversations")
	}

	s.logger.Info("Found participants for user",
		"user_id", userID,
		"user_type", userType,
		"user_type_string", string(userType),
		"participant_count", len(participants))

	// Log participant details for debugging
	for i, participant := range participants {
		s.logger.Debug("Participant details",
			"index", i,
			"participant_id", participant.ID,
			"conversation_id", participant.ConversationID,
			"participant_user_id", participant.UserID,
			"participant_user_type", participant.UserType,
			"participant_user_type_string", string(participant.UserType),
			"is_active", participant.IsActive)
	}

	var conversations []dto.ConversationResponse

	for _, participant := range participants {
		conversation, errFind := conversationRepo.FindByID(ctx, participant.ConversationID)
		if errFind != nil {
			s.logger.Errorf(
				"Failed to get conversation %s: %v",
				participant.ConversationID,
				errFind,
			)

			continue
		}

		if conversation != nil {
			conversations = append(conversations, *mapper.MapToConversationResponse(conversation))
		}
	}

	s.logger.Info("Returning conversations for user",
		"user_id", userID,
		"user_type", userType,
		"conversation_count", len(conversations))

	return conversations, nil
}

// GetUserConversationsWithCursor retrieves all conversations for a user using cursor-based pagination.
func (s *chatService) GetUserConversationsWithCursor(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
	limit int,
	afterCursor string,
	beforeCursor string,
) ([]dto.ConversationResponse, *pkgdto.CursorPagination, error) {
	conversationRepo := s.dataStore.ConversationRepository()

	// Get conversations with cursor
	conversations, err := conversationRepo.FindByUserIDWithCursor(
		ctx,
		userID,
		userType,
		limit,
		afterCursor,
		beforeCursor,
	)
	if err != nil {
		s.logger.Error("Failed to find conversations with cursor",
			"user_id", userID,
			"user_type", userType,
			"error", err)

		return nil, nil, httperror.NewInternalServerError("failed to get user conversations")
	}

	// Check if there are more results
	hasMore := len(conversations) > limit
	if hasMore {
		conversations = conversations[:limit]
	}

	// Map to response
	var conversationResponses []dto.ConversationResponse
	for _, conv := range conversations {
		conversationResponses = append(
			conversationResponses,
			*mapper.MapToConversationResponse(conv),
		)
	}

	// Build cursor pagination
	paging := s.buildConversationCursorPagination(
		conversations,
		afterCursor,
		beforeCursor,
		hasMore,
		limit,
	)

	return conversationResponses, paging, nil
}

// buildConversationCursorPagination constructs cursor pagination data for conversations.
func (s *chatService) buildConversationCursorPagination(
	conversations []*entity.Conversation,
	afterCursor string,
	beforeCursor string,
	hasMore bool,
	limit int,
) *pkgdto.CursorPagination {
	var (
		nextCursor, prevCursor string
		hasNext, hasPrev       bool
	)

	if len(conversations) == 0 {
		return &pkgdto.CursorPagination{
			Limit: int64(limit),
		}
	}

	// Forward pagination
	if afterCursor != "" || (afterCursor == "" && beforeCursor == "") {
		hasNext = hasMore
		hasPrev = afterCursor != ""
		lastConv := conversations[len(conversations)-1]
		nextCursor = lastConv.ID.String()

		if afterCursor != "" {
			prevCursor = afterCursor
		}

		return &pkgdto.CursorPagination{
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			Limit:      int64(limit),
		}
	}

	// Backward pagination
	hasPrev = hasMore
	hasNext = true
	firstConv := conversations[0]
	prevCursor = firstConv.ID.String()
	nextCursor = beforeCursor

	return &pkgdto.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      int64(limit),
	}
}

// UpdateConversationStatus updates the status of a conversation.
func (s *chatService) UpdateConversationStatus(
	ctx context.Context,
	conversationID uuid.UUID,
	req *dto.UpdateConversationStatusRequest,
) (*dto.ConversationResponse, error) {
	conversationRepo := s.dataStore.ConversationRepository()

	conversation, err := conversationRepo.FindByID(ctx, conversationID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get conversation")
	}

	if conversation == nil {
		return nil, httperror.NewBadRequestError("conversation not found")
	}

	// Update status
	if err = conversation.UpdateStatus(req.Status); err != nil {
		return nil, httperror.NewBadRequestError(fmt.Sprintf("failed to update status: %v", err))
	}

	// Save updated conversation
	updatedConversation, err := conversationRepo.Update(ctx, conversation)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to update conversation")
	}

	// Broadcast status update to participants
	s.broadcastSystemMessage(
		conversationID,
		fmt.Sprintf("Conversation status updated to %s", req.Status),
	)

	return mapper.MapToConversationResponse(updatedConversation), nil
}

// EndConversation ends a conversation.
func (s *chatService) EndConversation(
	ctx context.Context,
	conversationID uuid.UUID,
) (*dto.ConversationResponse, error) {
	return s.UpdateConversationStatus(ctx, conversationID, &dto.UpdateConversationStatusRequest{
		Status: constant.ConversationStatusEnded,
	})
}

// GetConversationMessages retrieves messages for a conversation with pagination.
func (s *chatService) GetConversationMessages(
	ctx context.Context,
	conversationID uuid.UUID,
	userID uuid.UUID,
	limit, offset int,
) ([]dto.MessageResponse, *pkgdto.OffsetPagination, error) {
	messageRepo := s.dataStore.MessageRepository()

	// Verify user is participant
	if err := s.conversationAccess.VerifyUserAccess(ctx, conversationID, userID); err != nil {
		return nil, nil, err
	}

	// Get messages
	messages, err := messageRepo.FindByConversationID(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to get messages")
	}

	// Get total count
	totalCount, err := messageRepo.CountByConversationID(ctx, conversationID)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to count messages")
	}

	// Map to response
	var messageResponses []dto.MessageResponse
	for _, msg := range messages {
		messageResponses = append(messageResponses, *mapper.MapToMessageResponse(msg))
	}

	totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	page := int64((offset / limit) + 1)

	paging := &pkgdto.OffsetPagination{
		Page:      page,
		Size:      int64(limit),
		TotalItem: totalCount,
		TotalPage: totalPages,
	}

	return messageResponses, paging, nil
}

// GetConversationMessagesWithCursor retrieves messages for a conversation using cursor-based pagination.
func (s *chatService) GetConversationMessagesWithCursor(
	ctx context.Context,
	conversationID uuid.UUID,
	userID uuid.UUID,
	limit int,
	afterCursor string,
	beforeCursor string,
) ([]dto.MessageResponse, *pkgdto.CursorPagination, error) {
	messageRepo := s.dataStore.MessageRepository()

	// Verify user is participant
	if err := s.conversationAccess.VerifyUserAccess(ctx, conversationID, userID); err != nil {
		return nil, nil, err
	}

	// Get messages with cursor
	messages, err := messageRepo.FindByConversationIDWithCursor(
		ctx,
		conversationID,
		limit,
		afterCursor,
		beforeCursor,
	)
	if err != nil {
		return nil, nil, httperror.NewInternalServerError("failed to get messages")
	}

	// Check if there are more results
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	// Map to response
	var messageResponses []dto.MessageResponse
	for _, msg := range messages {
		messageResponses = append(messageResponses, *mapper.MapToMessageResponse(msg))
	}

	// Build cursor pagination
	paging := s.buildMessageCursorPagination(messages, afterCursor, beforeCursor, hasMore, limit)

	return messageResponses, paging, nil
}

// buildMessageCursorPagination constructs cursor pagination data.
func (s *chatService) buildMessageCursorPagination(
	messages []*entity.Message,
	afterCursor string,
	beforeCursor string,
	hasMore bool,
	limit int,
) *pkgdto.CursorPagination {
	var (
		nextCursor, prevCursor string
		hasNext, hasPrev       bool
	)

	if len(messages) == 0 {
		return &pkgdto.CursorPagination{
			Limit: int64(limit),
		}
	}

	// Forward pagination
	if afterCursor != "" || (afterCursor == "" && beforeCursor == "") {
		hasNext = hasMore
		hasPrev = afterCursor != ""
		lastMsg := messages[len(messages)-1]
		nextCursor = lastMsg.ID.String()

		if afterCursor != "" {
			prevCursor = afterCursor
		}

		return &pkgdto.CursorPagination{
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			Limit:      int64(limit),
		}
	}

	// Backward pagination
	hasPrev = hasMore
	hasNext = true
	firstMsg := messages[0]
	prevCursor = firstMsg.ID.String()
	nextCursor = beforeCursor

	return &pkgdto.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Limit:      int64(limit),
	}
}

// GetMessageByID retrieves a message by ID without access control (for federation).
func (s *chatService) GetMessageByID(
	ctx context.Context,
	messageID uuid.UUID,
) (*dto.MessageResponse, error) {
	messageRepo := s.dataStore.MessageRepository()

	message, err := messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get message")
	}

	if message == nil {
		return nil, httperror.NewBadRequestError("message not found")
	}

	return mapper.MapToMessageResponse(message), nil
}

// JoinConversation adds a participant to a conversation.
func (s *chatService) JoinConversation(
	ctx context.Context,
	conversationID uuid.UUID,
	userID uuid.UUID,
	userType constant.UserType,
	role constant.ParticipantRole,
) (*dto.ParticipantResponse, error) {
	participantRepo := s.dataStore.ParticipantRepository()

	// Check if user is already a participant
	existingParticipants, err := participantRepo.FindByConversationID(ctx, conversationID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to check existing participants")
	}

	for _, p := range existingParticipants {
		if p.UserID == userID && p.IsActive {
			return nil, httperror.NewBadRequestError("user is already a participant")
		}
	}

	// Create participant
	participant, err := entity.NewParticipant(conversationID, userID, userType, role)
	if err != nil {
		return nil, httperror.NewBadRequestError(
			fmt.Sprintf("failed to create participant: %v", err),
		)
	}

	// Save participant
	savedParticipant, err := participantRepo.Create(ctx, participant)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to add participant")
	}

	// Broadcast join event
	s.broadcastSystemMessage(conversationID, "User joined the conversation")

	s.logger.Infof("User %s joined conversation %s", userID, conversationID)

	return mapper.MapToParticipantResponse(savedParticipant), nil
}

// LeaveConversation removes a participant from a conversation.
func (s *chatService) LeaveConversation(
	ctx context.Context,
	conversationID uuid.UUID,
	userID uuid.UUID,
) error {
	participantRepo := s.dataStore.ParticipantRepository()

	// Find participant
	participants, err := participantRepo.FindActiveByConversationID(ctx, conversationID)
	if err != nil {
		return httperror.NewInternalServerError("failed to find participants")
	}

	var participantID uuid.UUID

	found := false

	for _, p := range participants {
		if p.UserID == userID && p.IsActive {
			participantID = p.ID
			found = true

			break
		}
	}

	if !found {
		return httperror.NewBadRequestError("user is not an active participant")
	}

	// Mark as left
	if err = participantRepo.MarkAsLeft(ctx, participantID); err != nil {
		return httperror.NewInternalServerError("failed to leave conversation")
	}

	// Broadcast leave event
	s.broadcastSystemMessage(conversationID, "User left the conversation")

	s.logger.Infof("User %s left conversation %s", userID, conversationID)

	return nil
}

// GetConversationParticipants retrieves all participants for a conversation.
func (s *chatService) GetConversationParticipants(
	ctx context.Context,
	conversationID uuid.UUID,
) ([]dto.ParticipantResponse, error) {
	participantRepo := s.dataStore.ParticipantRepository()

	participants, err := participantRepo.FindByConversationID(ctx, conversationID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get participants")
	}

	var responses []dto.ParticipantResponse
	for _, p := range participants {
		responses = append(responses, *mapper.MapToParticipantResponse(p))
	}

	return responses, nil
}

// AssignConversationToAdmin assigns a conversation to an admin.
func (s *chatService) AssignConversationToAdmin(
	ctx context.Context,
	conversationID uuid.UUID,
	adminID uuid.UUID,
) (*dto.ConversationResponse, error) {
	_, err := s.JoinConversation(
		ctx,
		conversationID,
		adminID,
		constant.UserTypeAdmin,
		constant.ParticipantRoleModerator,
	)
	if err != nil {
		return nil, err
	}

	// Update conversation status to active
	_, err = s.UpdateConversationStatus(
		ctx,
		conversationID,
		&dto.UpdateConversationStatusRequest{
			Status: constant.ConversationStatusActive,
		},
	)
	if err != nil {
		return nil, err
	}

	// Get updated conversation
	return s.GetConversation(ctx, conversationID, adminID)
}

// GetWaitingConversations retrieves conversations waiting for admin assignment.
func (s *chatService) GetWaitingConversations(
	ctx context.Context,
) ([]dto.ConversationResponse, error) {
	conversationRepo := s.dataStore.ConversationRepository()

	// Get first 50 waiting conversations
	conversations, err := conversationRepo.FindByStatus(
		ctx,
		constant.ConversationStatusWaiting,
		constant.DefaultMessageLimit,
		0,
	)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get waiting conversations")
	}

	var responses []dto.ConversationResponse
	for _, c := range conversations {
		responses = append(responses, *mapper.MapToConversationResponse(c))
	}

	return responses, nil
}

// NOTE: broadcastMessage removed - broadcasting now handled directly in WebSocket handlers

// broadcastSystemMessage broadcasts a system message to conversation participants.
func (s *chatService) broadcastSystemMessage(
	conversationID uuid.UUID,
	content string,
) {
	if s.hub == nil {
		return
	}

	wsMessage, err := websocket.NewSystemMessage(content, nil)
	if err != nil {
		s.logger.Errorf("Failed to create system message: %v", err)
		return
	}

	err = s.hub.BroadcastToConversation(conversationID, wsMessage)
	if err != nil {
		s.logger.Errorf("Failed to broadcast system message: %v", err)
		return
	}
}
