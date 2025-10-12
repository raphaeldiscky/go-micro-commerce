package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/notification-service/internal/entity"
)

// NotificationRepository defines the methods for interacting with notifications.
type NotificationRepository interface {
	// Create inserts a new notification
	Create(ctx context.Context, notification *entity.Notification) (*entity.Notification, error)

	// FindByID retrieves a notification by ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error)

	// FindByUserIDWithCursor retrieves notifications for a user with cursor-based pagination
	FindByUserIDWithCursor(
		ctx context.Context,
		userID uuid.UUID,
		limit int64,
		cursorID string,
		cursorTimestamp int64,
	) ([]*entity.Notification, error)

	// FindUnreadByUserIDWithCursor retrieves unread notifications for a user with cursor-based pagination
	FindUnreadByUserIDWithCursor(
		ctx context.Context,
		userID uuid.UUID,
		limit int64,
		cursorID string,
		cursorTimestamp int64,
	) ([]*entity.Notification, error)

	// CountByUserID counts total notifications for a user
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// CountUnreadByUserID counts unread notifications for a user
	CountUnreadByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// MarkAsRead marks a notification as read
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// MarkAllAsRead marks all notifications as read for a user
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
}

// notificationRepository implements the NotificationRepository.
type notificationRepository struct {
	db DBTX
}

// NewNotificationRepository creates a new instance of notificationRepository.
func NewNotificationRepository(db DBTX) NotificationRepository {
	return &notificationRepository{
		db: db,
	}
}

// Create inserts a new notification.
func (r *notificationRepository) Create(
	ctx context.Context,
	notification *entity.Notification,
) (*entity.Notification, error) {
	query := `
		INSERT INTO notifications (
			id, user_id, type, title, message, metadata, is_read, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, type, title, message, metadata, is_read, read_at, created_at, updated_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.Metadata,
		notification.IsRead,
		notification.CreatedAt,
		notification.UpdatedAt,
	)

	createdNotification, err := scanNotification(row)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return createdNotification, nil
}

// FindByID retrieves a notification by ID.
func (r *notificationRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, metadata, is_read, read_at, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	return scanNotification(row)
}

// FindByUserIDWithCursor retrieves notifications for a user with cursor-based pagination.
func (r *notificationRepository) FindByUserIDWithCursor(
	ctx context.Context,
	userID uuid.UUID,
	limit int64,
	cursorID string,
	cursorTimestamp int64,
) ([]*entity.Notification, error) {
	var (
		query string
		args  []any
	)

	if cursorID != "" && cursorTimestamp > 0 {
		// Forward pagination with cursor
		query = `
			SELECT id, user_id, type, title, message, metadata, is_read, read_at, created_at, updated_at
			FROM notifications
			WHERE user_id = $1
				AND (
					EXTRACT(EPOCH FROM created_at)::BIGINT < $2
					OR (EXTRACT(EPOCH FROM created_at)::BIGINT = $2 AND id::text < $3)
				)
			ORDER BY created_at DESC, id DESC
			LIMIT $4
		`
		args = []any{userID, cursorTimestamp, cursorID, limit}
	} else {
		// No cursor: get first page
		query = `
			SELECT id, user_id, type, title, message, metadata, is_read, read_at, created_at, updated_at
			FROM notifications
			WHERE user_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		`
		args = []any{userID, limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications with cursor: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows)
}

// FindUnreadByUserIDWithCursor retrieves unread notifications for a user with cursor-based pagination.
func (r *notificationRepository) FindUnreadByUserIDWithCursor(
	ctx context.Context,
	userID uuid.UUID,
	limit int64,
	cursorID string,
	cursorTimestamp int64,
) ([]*entity.Notification, error) {
	var (
		query string
		args  []any
	)

	if cursorID != "" && cursorTimestamp > 0 {
		// Forward pagination with cursor
		query = `
			SELECT id, user_id, type, title, message, metadata, is_read, read_at, created_at, updated_at
			FROM notifications
			WHERE user_id = $1 AND is_read = false
				AND (
					EXTRACT(EPOCH FROM created_at)::BIGINT < $2
					OR (EXTRACT(EPOCH FROM created_at)::BIGINT = $2 AND id::text < $3)
				)
			ORDER BY created_at DESC, id DESC
			LIMIT $4
		`
		args = []any{userID, cursorTimestamp, cursorID, limit}
	} else {
		// No cursor: get first page of unread notifications
		query = `
			SELECT id, user_id, type, title, message, metadata, is_read, read_at, created_at, updated_at
			FROM notifications
			WHERE user_id = $1 AND is_read = false
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		`
		args = []any{userID, limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query unread notifications with cursor: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows)
}

// CountByUserID counts total notifications for a user.
func (r *notificationRepository) CountByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (int64, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`

	var count int64

	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	return count, nil
}

// CountUnreadByUserID counts unread notifications for a user.
func (r *notificationRepository) CountUnreadByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (int64, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`

	var count int64

	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

// MarkAsRead marks a notification as read.
func (r *notificationRepository) MarkAsRead(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND is_read = false
	`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("notification not found or already read")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user.
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND is_read = false
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// scanNotification scans a database row into a Notification struct.
func scanNotification(row pgx.Row) (*entity.Notification, error) {
	var notification entity.Notification

	var readAt pgtype.Timestamptz

	err := row.Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.Metadata,
		&notification.IsRead,
		&readAt,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("notification not found")
		}

		return nil, err
	}

	// Handle nullable read_at
	if readAt.Status == pgtype.Present {
		notification.ReadAt = &readAt.Time
	}

	return &notification, nil
}

// scanNotifications scans multiple rows into a slice of Notification structs.
func scanNotifications(rows pgx.Rows) ([]*entity.Notification, error) {
	var notifications []*entity.Notification

	for rows.Next() {
		notification, err := scanNotification(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}
