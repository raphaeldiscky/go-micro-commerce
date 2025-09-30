package mapper

import (
	pkgdto "github.com/raphaeldiscky/go-micro-commerce/pkg/dto"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/graph"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/dto"
)

// MapConversationToGraphQL maps a ConversationResponse to a Conversation GraphQL type.
func MapConversationToGraphQL(conv *dto.ConversationResponse) *graph.Conversation {
	return &graph.Conversation{
		ID:        conv.ID.String(),
		Subject:   conv.Subject,
		Status:    conv.Status,
		Priority:  conv.Priority,
		CreatedAt: conv.CreatedAt,
		UpdatedAt: conv.UpdatedAt,
		EndedAt:   conv.EndedAt,
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
		MessageType:    msg.MessageType,
		IsSystem:       msg.IsSystem,
		CreatedAt:      msg.CreatedAt,
	}
}

// MapParticipantToGraphQL maps a ParticipantResponse to a Participant GraphQL type.
func MapParticipantToGraphQL(p *dto.ParticipantResponse) *graph.Participant {
	return &graph.Participant{
		ID:             p.ID.String(),
		ConversationID: p.ConversationID.String(),
		UserID:         p.UserID.String(),
		UserType:       p.UserType,
		Role:           p.Role,
		JoinedAt:       p.JoinedAt,
		LeftAt:         p.LeftAt,
		IsActive:       p.IsActive,
	}
}

// MapMessagesToConnection maps messages and pagination to a MessageConnection.
func MapMessagesToConnection(
	messages []dto.MessageResponse,
	paging *pkgdto.OffsetPagination,
) *graph.MessageConnection {
	items := make([]*graph.Message, len(messages))
	for i, msg := range messages {
		items[i] = MapMessageToGraphQL(&msg)
	}

	return &graph.MessageConnection{
		Items: items,
		Pagination: &graph.OffsetPagination{
			TotalItems:  int(paging.TotalItem),
			TotalPages:  int(paging.TotalPage),
			CurrentPage: int(paging.Page),
			PageSize:    int(paging.Size),
			HasNext:     paging.Page < paging.TotalPage,
			HasPrev:     paging.Page > 1,
		},
	}
}
