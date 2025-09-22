package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/entity"
)

// ParticipantRepository defines the interface for participant data operations.
type ParticipantRepository interface {
	// Create creates a new participant
	Create(ctx context.Context, participant *entity.Participant) (*entity.Participant, error)

	// FindByID retrieves a participant by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Participant, error)

	// FindByConversationID retrieves all participants for a conversation
	FindByConversationID(
		ctx context.Context,
		conversationID uuid.UUID,
	) ([]*entity.Participant, error)

	// FindActiveByConversationID retrieves active participants for a conversation
	FindActiveByConversationID(
		ctx context.Context,
		conversationID uuid.UUID,
	) ([]*entity.Participant, error)

	// FindByUserID retrieves all participants for a user
	FindByUserID(
		ctx context.Context,
		userID uuid.UUID,
		userType constant.UserType,
	) ([]*entity.Participant, error)

	// FindActiveByUserID retrieves active participants for a user
	FindActiveByUserID(
		ctx context.Context,
		userID uuid.UUID,
		userType constant.UserType,
	) ([]*entity.Participant, error)

	// Update updates an existing participant
	Update(ctx context.Context, participant *entity.Participant) (*entity.Participant, error)

	// MarkAsLeft marks a participant as having left the conversation
	MarkAsLeft(ctx context.Context, id uuid.UUID) error

	// Delete removes a participant
	Delete(ctx context.Context, id uuid.UUID) error
}

// participantRepository implements the ParticipantRepository interface for PostgreSQL.
type participantRepository struct {
	db DBTX
}

// NewParticipantRepository creates a new instance of participantRepository.
func NewParticipantRepository(db DBTX) ParticipantRepository {
	return &participantRepository{
		db: db,
	}
}

// Create creates a new participant.
func (r *participantRepository) Create(
	ctx context.Context,
	participant *entity.Participant,
) (*entity.Participant, error) {
	query := `
		INSERT INTO participants (
			id, conversation_id, user_id, user_type, role, joined_at, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, conversation_id, user_id, user_type, role, joined_at, left_at, is_active
	`

	row := r.db.QueryRow(
		ctx,
		query,
		participant.ID,
		participant.ConversationID,
		participant.UserID,
		participant.UserType,
		participant.Role,
		participant.JoinedAt,
		participant.IsActive,
	)

	return r.scanParticipant(row)
}

// FindByID retrieves a participant by its ID.
func (r *participantRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Participant, error) {
	query := `
		SELECT id, conversation_id, user_id, user_type, role, joined_at, left_at, is_active
		FROM participants
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	return r.scanParticipant(row)
}

// FindByConversationID retrieves all participants for a conversation.
func (r *participantRepository) FindByConversationID(
	ctx context.Context,
	conversationID uuid.UUID,
) ([]*entity.Participant, error) {
	query := `
		SELECT id, conversation_id, user_id, user_type, role, joined_at, left_at, is_active
		FROM participants
		WHERE conversation_id = $1
		ORDER BY joined_at ASC
	`

	rows, err := r.db.Query(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query participants by conversation: %w", err)
	}
	defer rows.Close()

	return r.scanParticipants(rows)
}

// FindActiveByConversationID retrieves active participants for a conversation.
func (r *participantRepository) FindActiveByConversationID(
	ctx context.Context,
	conversationID uuid.UUID,
) ([]*entity.Participant, error) {
	query := `
		SELECT id, conversation_id, user_id, user_type, role, joined_at, left_at, is_active
		FROM participants
		WHERE conversation_id = $1 AND is_active = TRUE
		ORDER BY joined_at ASC
	`

	rows, err := r.db.Query(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active participants: %w", err)
	}
	defer rows.Close()

	return r.scanParticipants(rows)
}

// FindByUserID retrieves all participants for a user.
func (r *participantRepository) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
) ([]*entity.Participant, error) {
	query := `
		SELECT id, conversation_id, user_id, user_type, role, joined_at, left_at, is_active
		FROM participants
		WHERE user_id = $1 AND user_type = $2
		ORDER BY joined_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID, userType)
	if err != nil {
		return nil, fmt.Errorf("failed to query participants by user: %w", err)
	}
	defer rows.Close()

	return r.scanParticipants(rows)
}

// FindActiveByUserID retrieves active participants for a user.
func (r *participantRepository) FindActiveByUserID(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
) ([]*entity.Participant, error) {
	query := `
		SELECT id, conversation_id, user_id, user_type, role, joined_at, left_at, is_active
		FROM participants
		WHERE user_id = $1 AND user_type = $2 AND is_active = TRUE
		ORDER BY joined_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID, userType)
	if err != nil {
		return nil, fmt.Errorf("failed to query active participants by user: %w", err)
	}
	defer rows.Close()

	return r.scanParticipants(rows)
}

// Update updates an existing participant.
func (r *participantRepository) Update(
	ctx context.Context,
	participant *entity.Participant,
) (*entity.Participant, error) {
	query := `
		UPDATE participants
		SET role = $2,
			left_at = $3,
			is_active = $4
		WHERE id = $1
		RETURNING id, conversation_id, user_id, user_type, role, joined_at, left_at, is_active
	`

	row := r.db.QueryRow(
		ctx,
		query,
		participant.ID,
		participant.Role,
		participant.LeftAt,
		participant.IsActive,
	)

	return r.scanParticipant(row)
}

// MarkAsLeft marks a participant as having left the conversation.
func (r *participantRepository) MarkAsLeft(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE participants
		SET left_at = NOW(),
			is_active = FALSE
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark participant as left: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("participant with id %s not found", id)
	}

	return nil
}

// Delete removes a participant.
func (r *participantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM participants WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete participant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("participant with id %s not found", id)
	}

	return nil
}

// scanParticipant scans a row into a Participant entity.
func (r *participantRepository) scanParticipant(row pgx.Row) (*entity.Participant, error) {
	var participant entity.Participant

	err := row.Scan(
		&participant.ID,
		&participant.ConversationID,
		&participant.UserID,
		&participant.UserType,
		&participant.Role,
		&participant.JoinedAt,
		&participant.LeftAt,
		&participant.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("participant not found")
		}

		return nil, fmt.Errorf("failed to scan participant: %w", err)
	}

	return &participant, nil
}

// scanParticipants scans multiple rows into Participant entities.
func (r *participantRepository) scanParticipants(rows pgx.Rows) ([]*entity.Participant, error) {
	var participants []*entity.Participant

	for rows.Next() {
		var participant entity.Participant

		err := rows.Scan(
			&participant.ID,
			&participant.ConversationID,
			&participant.UserID,
			&participant.UserType,
			&participant.Role,
			&participant.JoinedAt,
			&participant.LeftAt,
			&participant.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		participants = append(participants, &participant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return participants, nil
}
