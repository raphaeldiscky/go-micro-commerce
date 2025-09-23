// Package service provides business logic for chat operations.
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

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
	GetUserConversations(
		ctx context.Context,
		userID uuid.UUID,
		userType constant.UserType,
	) ([]dto.ConversationResponse, error)
	UpdateConversationStatus(
		ctx context.Context,
		conversationID uuid.UUID,
		req *dto.UpdateConversationStatusRequest,
	) (*dto.ConversationResponse, error)
	SetConversationSubject(
		ctx context.Context,
		conversationID uuid.UUID,
		req *dto.SetConversationSubjectRequest,
	) (*dto.ConversationResponse, error)
	EndConversation(
		ctx context.Context,
		conversationID uuid.UUID,
	) (*dto.ConversationResponse, error)

	// Message management
	SendMessage(
		ctx context.Context,
		req *dto.CreateMessageRequest,
		senderID uuid.UUID,
	) (*dto.MessageResponse, error)
	GetConversationMessages(
		ctx context.Context,
		conversationID uuid.UUID,
		userID uuid.UUID,
		limit, offset int,
	) (*dto.MessageListResponse, error)

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

// GetUserConversations retrieves all conversations for a user.
func (s *chatService) GetUserConversations(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
) ([]dto.ConversationResponse, error) {
	participantRepo := s.dataStore.ParticipantRepository()
	conversationRepo := s.dataStore.ConversationRepository()

	// Get user's active participations
	participants, err := participantRepo.FindActiveByUserID(ctx, userID, userType)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get user conversations")
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

	return conversations, nil
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

// SetConversationSubject sets the subject of a conversation.
func (s *chatService) SetConversationSubject(
	ctx context.Context,
	conversationID uuid.UUID,
	req *dto.SetConversationSubjectRequest,
) (*dto.ConversationResponse, error) {
	conversationRepo := s.dataStore.ConversationRepository()

	conversation, err := conversationRepo.FindByID(ctx, conversationID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get conversation")
	}

	if conversation == nil {
		return nil, httperror.NewBadRequestError("conversation not found")
	}

	// Update subject
	conversation.SetSubject(req.Subject)

	// Save updated conversation
	updatedConversation, err := conversationRepo.Update(ctx, conversation)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to update conversation")
	}

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

// SendMessage sends a message to a conversation.
func (s *chatService) SendMessage(
	ctx context.Context,
	req *dto.CreateMessageRequest,
	senderID uuid.UUID,
) (*dto.MessageResponse, error) {
	var result *dto.MessageResponse

	err := s.dataStore.Atomic(ctx, func(ds repository.DataStore) error {
		messageRepo := ds.MessageRepository()

		// Verify sender is participant
		if err := s.conversationAccess.VerifyActiveUserAccess(ctx, req.ConversationID, senderID); err != nil {
			return err
		}

		// Create message entity
		message, err := entity.NewMessage(
			req.ConversationID,
			senderID,
			req.Content,
			req.MessageType,
		)
		if err != nil {
			return httperror.NewBadRequestError(fmt.Sprintf("failed to create message: %v", err))
		}

		// Add metadata if provided
		if req.Metadata != nil {
			message.Metadata = req.Metadata
		}

		// Save message
		savedMessage, err := messageRepo.Create(ctx, message)
		if err != nil {
			return httperror.NewInternalServerError("failed to save message")
		}

		result = mapper.MapToMessageResponse(savedMessage)

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Broadcast message to conversation participants via WebSocket
	s.broadcastMessage(req.ConversationID, result)

	s.logger.Infof("Message sent by user %s to conversation %s", senderID, req.ConversationID)

	return result, nil
}

// GetConversationMessages retrieves messages for a conversation with pagination.
func (s *chatService) GetConversationMessages(
	ctx context.Context,
	conversationID uuid.UUID,
	userID uuid.UUID,
	limit, offset int,
) (*dto.MessageListResponse, error) {
	messageRepo := s.dataStore.MessageRepository()

	// Verify user is participant
	if err := s.conversationAccess.VerifyUserAccess(ctx, conversationID, userID); err != nil {
		return nil, err
	}

	// Get messages
	messages, err := messageRepo.FindByConversationID(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to get messages")
	}

	// Get total count
	totalCount, err := messageRepo.CountByConversationID(ctx, conversationID)
	if err != nil {
		return nil, httperror.NewInternalServerError("failed to count messages")
	}

	// Map to response
	var messageResponses []dto.MessageResponse
	for _, msg := range messages {
		messageResponses = append(messageResponses, *mapper.MapToMessageResponse(msg))
	}

	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))
	page := (offset / limit) + 1

	return &dto.MessageListResponse{
		Messages:   messageResponses,
		Total:      totalCount,
		Page:       page,
		PerPage:    limit,
		TotalPages: totalPages,
	}, nil
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

// broadcastMessage broadcasts a message to conversation participants via WebSocket.
func (s *chatService) broadcastMessage(
	conversationID uuid.UUID,
	message *dto.MessageResponse,
) {
	if s.hub == nil {
		return
	}

	wsMessage, err := websocket.NewChatMessage(
		conversationID,
		*message.SenderID,
		message.Content,
		message.MessageType,
	)
	if err != nil {
		s.logger.Errorf("Failed to create WebSocket message: %v", err)
		return
	}

	err = s.hub.BroadcastToConversation(conversationID, wsMessage)
	if err != nil {
		s.logger.Errorf("Failed to broadcast message: %v", err)
		return
	}
}

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
