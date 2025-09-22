// Package mapper provides functions for mapping chat entities to DTOs.
package mapper

import (
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/entity"
)

// MapToConversationResponse converts conversation entity to DTO response.
func MapToConversationResponse(conversation *entity.Conversation) *dto.ConversationResponse {
	return &dto.ConversationResponse{
		ID:        conversation.ID,
		Status:    conversation.Status,
		Subject:   conversation.Subject,
		Priority:  conversation.Priority,
		Metadata:  conversation.Metadata,
		CreatedAt: conversation.CreatedAt,
		UpdatedAt: conversation.UpdatedAt,
		EndedAt:   conversation.EndedAt,
	}
}

// MapToMessageResponse converts message entity to DTO response.
func MapToMessageResponse(message *entity.Message) *dto.MessageResponse {
	return &dto.MessageResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		SenderID:       message.SenderID,
		Content:        message.Content,
		MessageType:    message.MessageType,
		Metadata:       message.Metadata,
		IsSystem:       message.IsSystem,
		CreatedAt:      message.CreatedAt,
	}
}

// MapToParticipantResponse converts participant entity to DTO response.
func MapToParticipantResponse(participant *entity.Participant) *dto.ParticipantResponse {
	return &dto.ParticipantResponse{
		ID:             participant.ID,
		ConversationID: participant.ConversationID,
		UserID:         participant.UserID,
		UserType:       participant.UserType,
		Role:           participant.Role,
		JoinedAt:       participant.JoinedAt,
		LeftAt:         participant.LeftAt,
		IsActive:       participant.IsActive,
	}
}
