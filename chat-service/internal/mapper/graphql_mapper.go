package mapper

import (
	"strings"

	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/dto"
)

// MapConversationToGraphQL maps a ConversationResponse to a Conversation GraphQL type.
func MapConversationToGraphQL(conv *dto.ConversationResponse) *graph.Conversation {
	return &graph.Conversation{
		ID:               conv.ID.String(),
		Subject:          conv.Subject,
		Status:           constant.ConversationStatus(strings.ToUpper(string(conv.Status))),
		Priority:         conv.Priority,
		ParticipantCount: conv.ParticipantCount,
		CreatedAt:        conv.CreatedAt,
		UpdatedAt:        conv.UpdatedAt,
		EndedAt:          conv.EndedAt,
	}
}

// MapMessageToGraphQL maps a MessageResponse to a Message GraphQL type.
func MapMessageToGraphQL(msg *dto.MessageResponse) *graph.Message {
	var senderID *string

	if msg.SenderID != nil {
		id := msg.SenderID.String()
		senderID = &id
	}

	return &graph.Message{
		ID:             msg.ID.String(),
		ConversationID: msg.ConversationID.String(),
		SenderID:       senderID,
		Content:        msg.Content,
		MessageType:    constant.MessageType(strings.ToUpper(string(msg.MessageType))),
		IsSystem:       msg.IsSystem,
		CreatedAt:      msg.CreatedAt,
	}
}

// convertParticipantRoleToGraphQL converts backend ParticipantRole to GraphQL enum.
func convertParticipantRoleToGraphQL(role constant.ParticipantRole) constant.ParticipantRole {
	// Backend uses: participant, moderator, observer
	// GraphQL uses: MEMBER, OWNER, MODERATOR
	switch role {
	case constant.ParticipantRoleParticipant:
		return "MEMBER"
	case constant.ParticipantRoleModerator:
		return "MODERATOR"
	case constant.ParticipantRoleObserver:
		return "OWNER"
	default:
		return constant.ParticipantRole(strings.ToUpper(string(role)))
	}
}

// MapParticipantToGraphQL maps a ParticipantResponse to a Participant GraphQL type.
func MapParticipantToGraphQL(p *dto.ParticipantResponse) *graph.Participant {
	return &graph.Participant{
		ID:             p.ID.String(),
		ConversationID: p.ConversationID.String(),
		UserID:         p.UserID.String(),
		UserType:       constant.UserType(strings.ToUpper(string(p.UserType))),
		Role:           convertParticipantRoleToGraphQL(p.Role),
		JoinedAt:       p.JoinedAt,
		LeftAt:         p.LeftAt,
		IsActive:       p.IsActive,
	}
}

// MapMessagesToCursorConnection maps messages and cursor pagination to a MessageConnection.
func MapMessagesToCursorConnection(
	messages []dto.MessageResponse,
	paging *pkgdto.CursorPagination,
) *graph.MessageConnection {
	edges := make([]*graph.MessageEdge, len(messages))
	for i, msg := range messages {
		// Use message ID as cursor
		cursor := msg.ID.String()
		edges[i] = &graph.MessageEdge{
			Cursor: cursor,
			Node:   MapMessageToGraphQL(&msg),
		}
	}

	// Determine start and end cursors
	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &graph.MessageConnection{
		Edges: edges,
		PageInfo: &graph.PageInfo{
			HasNextPage:     paging.HasNext,
			HasPreviousPage: paging.HasPrev,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
	}
}
