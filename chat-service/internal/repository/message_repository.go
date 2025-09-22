package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/entity"
)

// MessageRepository defines the interface for message data operations.
type MessageRepository interface {
	// Create creates a new message
	Create(ctx context.Context, message *entity.Message) (*entity.Message, error)

	// FindByID retrieves a message by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Message, error)

	// FindByConversationID retrieves messages for a conversation with pagination
	FindByConversationID(
		ctx context.Context,
		conversationID uuid.UUID,
		limit, offset int,
	) ([]*entity.Message, error)

	// FindLatestByConversationID retrieves the latest messages for a conversation
	FindLatestByConversationID(
		ctx context.Context,
		conversationID uuid.UUID,
		limit int,
	) ([]*entity.Message, error)

	// CountByConversationID counts messages in a conversation
	CountByConversationID(ctx context.Context, conversationID uuid.UUID) (int64, error)
}

// messageRepository implements the MessageRepository interface for PostgreSQL.
type messageRepository struct {
	db DBTX
}

// NewMessageRepository creates a new instance of messageRepository.
func NewMessageRepository(db DBTX) MessageRepository {
	return &messageRepository{
		db: db,
	}
}

// Create creates a new message.
func (r *messageRepository) Create(
	ctx context.Context,
	message *entity.Message,
) (*entity.Message, error) {
	metadataJSON, err := json.Marshal(message.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO messages (
			id, conversation_id, sender_id, content,
			message_type, metadata, is_system, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, conversation_id, sender_id, content,
			message_type, metadata, is_system, created_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		message.ID,
		message.ConversationID,
		message.SenderID,
		message.Content,
		message.MessageType,
		metadataJSON,
		message.IsSystem,
		message.CreatedAt,
	)

	return r.scanMessage(row)
}

// FindByID retrieves a message by its ID.
func (r *messageRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, content,
			message_type, metadata, is_system, created_at
		FROM messages
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	return r.scanMessage(row)
}

// FindByConversationID retrieves messages for a conversation with pagination.
func (r *messageRepository) FindByConversationID(
	ctx context.Context,
	conversationID uuid.UUID,
	limit, offset int,
) ([]*entity.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, content,
			message_type, metadata, is_system, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages by conversation: %w", err)
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

// FindLatestByConversationID retrieves the latest messages for a conversation.
func (r *messageRepository) FindLatestByConversationID(
	ctx context.Context,
	conversationID uuid.UUID,
	limit int,
) ([]*entity.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, content,
			message_type, metadata, is_system, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest messages: %w", err)
	}
	defer rows.Close()

	messages, err := r.scanMessages(rows)
	if err != nil {
		return nil, err
	}

	// Reverse the order to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CountByConversationID counts messages in a conversation.
func (r *messageRepository) CountByConversationID(
	ctx context.Context,
	conversationID uuid.UUID,
) (int64, error) {
	query := `SELECT COUNT(*) FROM messages WHERE conversation_id = $1`

	var count int64

	err := r.db.QueryRow(ctx, query, conversationID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// scanMessage scans a row into a Message entity.
func (r *messageRepository) scanMessage(row pgx.Row) (*entity.Message, error) {
	var (
		message      entity.Message
		metadataJSON []byte
	)

	err := row.Scan(
		&message.ID,
		&message.ConversationID,
		&message.SenderID,
		&message.Content,
		&message.MessageType,
		&metadataJSON,
		&message.IsSystem,
		&message.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("message not found")
		}

		return nil, fmt.Errorf("failed to scan message: %w", err)
	}

	// Unmarshal metadata
	if metadataJSON != nil {
		if err = json.Unmarshal(metadataJSON, &message.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		message.Metadata = make(map[string]any)
	}

	return &message, nil
}

// scanMessages scans multiple rows into Message entities.
func (r *messageRepository) scanMessages(rows pgx.Rows) ([]*entity.Message, error) {
	var messages []*entity.Message

	for rows.Next() {
		var (
			message      entity.Message
			metadataJSON []byte
		)

		err := rows.Scan(
			&message.ID,
			&message.ConversationID,
			&message.SenderID,
			&message.Content,
			&message.MessageType,
			&metadataJSON,
			&message.IsSystem,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		// Unmarshal metadata
		if metadataJSON != nil {
			if err = json.Unmarshal(metadataJSON, &message.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		} else {
			message.Metadata = make(map[string]any)
		}

		messages = append(messages, &message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return messages, nil
}
