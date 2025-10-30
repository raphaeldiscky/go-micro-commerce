package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/entity"
)

// ConversationRepository defines the interface for conversation data operations.
type ConversationRepository interface {
	// Create creates a new conversation
	Create(ctx context.Context, conversation *entity.Conversation) (*entity.Conversation, error)

	// Update updates an existing conversation
	Update(ctx context.Context, conversation *entity.Conversation) (*entity.Conversation, error)

	// UpdateStatus updates only the status of a conversation
	UpdateStatus(ctx context.Context, id uuid.UUID, status constant.ConversationStatus) error

	// FindByID retrieves a conversation by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error)

	// FindByStatus retrieves conversations by status with pagination
	FindByStatus(
		ctx context.Context,
		status constant.ConversationStatus,
		limit, offset int,
	) ([]*entity.Conversation, error)

	// FindWaitingConversations retrieves conversations waiting for admin assignment
	FindWaitingConversations(ctx context.Context, limit int) ([]*entity.Conversation, error)

	// FindActiveByAdmin retrieves active conversations assigned to an admin
	FindActiveByAdmin(ctx context.Context, adminID uuid.UUID) ([]*entity.Conversation, error)

	// FindByUserIDWithCursor retrieves conversations for a user using cursor-based pagination
	FindByUserIDWithCursor(
		ctx context.Context,
		userID uuid.UUID,
		userType constant.UserType,
		limit int,
		afterCursor string,
		beforeCursor string,
	) ([]*entity.Conversation, error)

	// Delete soft deletes a conversation
	Delete(ctx context.Context, id uuid.UUID) error
}

// conversationRepository implements the ConversationRepository interface for PostgreSQL.
type conversationRepository struct {
	db DBTX
}

// NewConversationRepository creates a new instance of conversationRepository.
func NewConversationRepository(db DBTX) ConversationRepository {
	return &conversationRepository{
		db: db,
	}
}

// Create creates a new conversation.
func (r *conversationRepository) Create(
	ctx context.Context,
	conversation *entity.Conversation,
) (*entity.Conversation, error) {
	metadataJSON, err := sonic.Marshal(conversation.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO conversations (
			id, status, subject, priority, metadata, created_at, updated_at, ended_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, status, subject, priority, metadata, created_at, updated_at, ended_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		conversation.ID,
		conversation.Status,
		conversation.Subject,
		conversation.Priority,
		metadataJSON,
		conversation.CreatedAt,
		conversation.UpdatedAt,
		conversation.EndedAt,
	)

	return r.scanConversation(row)
}

// FindByID retrieves a conversation by its ID.
func (r *conversationRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Conversation, error) {
	query := `
		SELECT id, status, subject, priority, metadata, created_at, updated_at, ended_at
		FROM conversations
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	return r.scanConversation(row)
}

// FindByStatus retrieves conversations by status with pagination.
func (r *conversationRepository) FindByStatus(
	ctx context.Context,
	status constant.ConversationStatus,
	limit, offset int,
) ([]*entity.Conversation, error) {
	query := `
		SELECT id, status, subject, priority, metadata, created_at, updated_at, ended_at
		FROM conversations
		WHERE status = $1
		ORDER BY priority DESC, created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations by status: %w", err)
	}
	defer rows.Close()

	return r.scanConversations(rows)
}

// FindWaitingConversations retrieves conversations waiting for admin assignment.
func (r *conversationRepository) FindWaitingConversations(
	ctx context.Context,
	limit int,
) ([]*entity.Conversation, error) {
	query := `
		SELECT id, status, subject, priority, metadata, created_at, updated_at, ended_at
		FROM conversations
		WHERE status = $1
		ORDER BY priority DESC, created_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, constant.ConversationStatusWaiting, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query waiting conversations: %w", err)
	}
	defer rows.Close()

	return r.scanConversations(rows)
}

// FindActiveByAdmin retrieves active conversations assigned to an admin.
func (r *conversationRepository) FindActiveByAdmin(
	ctx context.Context,
	adminID uuid.UUID,
) ([]*entity.Conversation, error) {
	query := `
		SELECT c.id, c.status, c.subject, c.priority, c.metadata, c.created_at, c.updated_at, c.ended_at
		FROM conversations c
		INNER JOIN participants p ON c.id = p.conversation_id
		WHERE c.status = $1
		AND p.user_id = $2
		AND p.user_type = 'admin'
		AND p.is_active = TRUE
		ORDER BY c.updated_at DESC
	`

	rows, err := r.db.Query(ctx, query, constant.ConversationStatusActive, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to query admin conversations: %w", err)
	}
	defer rows.Close()

	return r.scanConversations(rows)
}

// Update updates an existing conversation.
func (r *conversationRepository) Update(
	ctx context.Context,
	conversation *entity.Conversation,
) (*entity.Conversation, error) {
	metadataJSON, err := sonic.Marshal(conversation.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE conversations
		SET status = $2,
			subject = $3,
			priority = $4,
			metadata = $5,
			updated_at = $6,
			ended_at = $7
		WHERE id = $1
		RETURNING id, status, subject, priority, metadata, created_at, updated_at, ended_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		conversation.ID,
		conversation.Status,
		conversation.Subject,
		conversation.Priority,
		metadataJSON,
		conversation.UpdatedAt,
		conversation.EndedAt,
	)

	return r.scanConversation(row)
}

// UpdateStatus updates only the status of a conversation.
func (r *conversationRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status constant.ConversationStatus,
) error {
	query := `
		UPDATE conversations
		SET status = $2,
			updated_at = NOW(),
			ended_at = CASE WHEN $2 = 'ended' AND ended_at IS NULL THEN NOW() ELSE ended_at END
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update conversation status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("conversation with id %s not found", id)
	}

	return nil
}

// FindByUserIDWithCursor retrieves conversations for a user using cursor-based pagination.
func (r *conversationRepository) FindByUserIDWithCursor(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
	limit int,
	afterCursor string,
	beforeCursor string,
) ([]*entity.Conversation, error) {
	var (
		query string
		args  []any
	)

	if afterCursor != "" && beforeCursor != "" {
		return nil, errors.New("cannot use both after and before cursors")
	}

	switch {
	case afterCursor != "":
		// Forward pagination: get conversations after the cursor
		query = `
			SELECT c.id, c.status, c.subject, c.priority, c.metadata, c.created_at, c.updated_at, c.ended_at,
			       COUNT(DISTINCT p2.id) FILTER (WHERE p2.is_active = TRUE) as participant_count
			FROM conversations c
			INNER JOIN participants p ON c.id = p.conversation_id
			LEFT JOIN participants p2 ON c.id = p2.conversation_id
			WHERE p.user_id = $1 AND p.user_type = $2 AND p.is_active = TRUE
				AND (c.updated_at, c.id) < (
					SELECT updated_at, id FROM conversations WHERE id = $3
				)
			GROUP BY c.id, c.status, c.subject, c.priority, c.metadata, c.created_at, c.updated_at, c.ended_at
			ORDER BY c.updated_at DESC, c.id DESC
			LIMIT $4
		`
		args = []any{userID, userType, afterCursor, limit + 1} // +1 to check hasNext
	case beforeCursor != "":
		// Backward pagination: get conversations before the cursor
		query = `
			SELECT c.id, c.status, c.subject, c.priority, c.metadata, c.created_at, c.updated_at, c.ended_at,
			       COUNT(DISTINCT p2.id) FILTER (WHERE p2.is_active = TRUE) as participant_count
			FROM conversations c
			INNER JOIN participants p ON c.id = p.conversation_id
			LEFT JOIN participants p2 ON c.id = p2.conversation_id
			WHERE p.user_id = $1 AND p.user_type = $2 AND p.is_active = TRUE
				AND (c.updated_at, c.id) > (
					SELECT updated_at, id FROM conversations WHERE id = $3
				)
			GROUP BY c.id, c.status, c.subject, c.priority, c.metadata, c.created_at, c.updated_at, c.ended_at
			ORDER BY c.updated_at ASC, c.id ASC
			LIMIT $4
		`
		args = []any{userID, userType, beforeCursor, limit + 1} // +1 to check hasPrev
	default:
		// No cursor: get first page
		query = `
			SELECT c.id, c.status, c.subject, c.priority, c.metadata, c.created_at, c.updated_at, c.ended_at,
			       COUNT(DISTINCT p2.id) FILTER (WHERE p2.is_active = TRUE) as participant_count
			FROM conversations c
			INNER JOIN participants p ON c.id = p.conversation_id
			LEFT JOIN participants p2 ON c.id = p2.conversation_id
			WHERE p.user_id = $1 AND p.user_type = $2 AND p.is_active = TRUE
			GROUP BY c.id, c.status, c.subject, c.priority, c.metadata, c.created_at, c.updated_at, c.ended_at
			ORDER BY c.updated_at DESC, c.id DESC
			LIMIT $3
		`
		args = []any{userID, userType, limit + 1} // +1 to check hasNext
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations with cursor: %w", err)
	}
	defer rows.Close()

	conversations, err := r.scanConversations(rows)
	if err != nil {
		return nil, err
	}

	// Reverse results for backward pagination
	if beforeCursor != "" && len(conversations) > 0 {
		for i, j := 0, len(conversations)-1; i < j; i, j = i+1, j-1 {
			conversations[i], conversations[j] = conversations[j], conversations[i]
		}
	}

	return conversations, nil
}

// Delete soft deletes a conversation by setting status to ended.
func (r *conversationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.UpdateStatus(ctx, id, constant.ConversationStatusEnded)
}

// scanConversation scans a row into a Conversation entity.
func (r *conversationRepository) scanConversation(row pgx.Row) (*entity.Conversation, error) {
	var (
		conversation entity.Conversation
		metadataJSON []byte
	)

	err := row.Scan(
		&conversation.ID,
		&conversation.Status,
		&conversation.Subject,
		&conversation.Priority,
		&metadataJSON,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
		&conversation.EndedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("conversation not found")
		}

		return nil, fmt.Errorf("failed to scan conversation: %w", err)
	}

	// Unmarshal metadata
	if metadataJSON != nil {
		if err = sonic.Unmarshal(metadataJSON, &conversation.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		conversation.Metadata = make(map[string]any)
	}

	return &conversation, nil
}

// scanConversations scans multiple rows into Conversation entities.
func (r *conversationRepository) scanConversations(rows pgx.Rows) ([]*entity.Conversation, error) {
	var conversations []*entity.Conversation

	for rows.Next() {
		var (
			conversation     entity.Conversation
			metadataJSON     []byte
			participantCount *int
		)

		err := rows.Scan(
			&conversation.ID,
			&conversation.Status,
			&conversation.Subject,
			&conversation.Priority,
			&metadataJSON,
			&conversation.CreatedAt,
			&conversation.UpdatedAt,
			&conversation.EndedAt,
			&participantCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		// Set participant count (default to 0 if NULL)
		if participantCount != nil {
			conversation.ParticipantCount = *participantCount
		}

		// Unmarshal metadata
		if metadataJSON != nil {
			if err = sonic.Unmarshal(metadataJSON, &conversation.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		} else {
			conversation.Metadata = make(map[string]any)
		}

		conversations = append(conversations, &conversation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return conversations, nil
}
