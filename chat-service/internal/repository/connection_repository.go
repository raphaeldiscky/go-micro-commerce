package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/chat-service/internal/entity"
)

// ConnectionRepository defines the interface for connection data operations.
type ConnectionRepository interface {
	// Create creates a new connection
	Create(ctx context.Context, connection *entity.Connection) (*entity.Connection, error)

	// FindByID retrieves a connection by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Connection, error)

	// FindByConnectionID retrieves a connection by its connection ID
	FindByConnectionID(ctx context.Context, connectionID string) (*entity.Connection, error)

	// FindByUserID retrieves all connections for a user
	FindByUserID(
		ctx context.Context,
		userID uuid.UUID,
		userType constant.UserType,
	) ([]*entity.Connection, error)

	// FindActiveByUserID retrieves active connections for a user
	FindActiveByUserID(
		ctx context.Context,
		userID uuid.UUID,
		userType constant.UserType,
	) ([]*entity.Connection, error)

	// FindAllActive retrieves all active connections
	FindAllActive(ctx context.Context) ([]*entity.Connection, error)

	// UpdateHeartbeat updates the last heartbeat timestamp
	UpdateHeartbeat(ctx context.Context, connectionID string) error

	// MarkAsInactive marks a connection as inactive
	MarkAsInactive(ctx context.Context, connectionID string) error

	// CleanupStaleConnections removes connections that haven't sent heartbeat recently
	CleanupStaleConnections(ctx context.Context, staleThreshold time.Duration) (int64, error)

	// Delete removes a connection
	Delete(ctx context.Context, id uuid.UUID) error
}

// connectionRepository implements the ConnectionRepository interface for PostgreSQL.
type connectionRepository struct {
	db DBTX
}

// NewConnectionRepository creates a new instance of connectionRepository.
func NewConnectionRepository(db DBTX) ConnectionRepository {
	return &connectionRepository{
		db: db,
	}
}

// Create creates a new connection.
func (r *connectionRepository) Create(
	ctx context.Context,
	connection *entity.Connection,
) (*entity.Connection, error) {
	query := `
		INSERT INTO connections (
			id, user_id, connection_id, socket_id,
			user_agent, ip_address, connected_at, last_heartbeat, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, connection_id, socket_id,
			user_agent, ip_address, connected_at, last_heartbeat, is_active
	`

	row := r.db.QueryRow(
		ctx,
		query,
		connection.ID,
		connection.UserID,
		connection.ConnectionID,
		connection.SocketID,
		connection.UserAgent,
		connection.IPAddress,
		connection.ConnectedAt,
		connection.LastHeartbeat,
		connection.IsActive,
	)

	return r.scanConnection(row)
}

// FindByID retrieves a connection by its ID.
func (r *connectionRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Connection, error) {
	query := `
		SELECT id, user_id, connection_id, socket_id,
			user_agent, ip_address, connected_at, last_heartbeat, is_active
		FROM connections
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	return r.scanConnection(row)
}

// FindByConnectionID retrieves a connection by its connection ID.
func (r *connectionRepository) FindByConnectionID(
	ctx context.Context,
	connectionID string,
) (*entity.Connection, error) {
	query := `
		SELECT id, user_id, connection_id, socket_id,
			user_agent, ip_address, connected_at, last_heartbeat, is_active
		FROM connections
		WHERE connection_id = $1
	`

	row := r.db.QueryRow(ctx, query, connectionID)

	return r.scanConnection(row)
}

// FindByUserID retrieves all connections for a user.
func (r *connectionRepository) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
) ([]*entity.Connection, error) {
	query := `
		SELECT id, user_id, connection_id, socket_id,
			user_agent, ip_address, connected_at, last_heartbeat, is_active
		FROM connections
		WHERE user_id = $1
		ORDER BY connected_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query connections by user: %w", err)
	}
	defer rows.Close()

	return r.scanConnections(rows)
}

// FindActiveByUserID retrieves active connections for a user.
func (r *connectionRepository) FindActiveByUserID(
	ctx context.Context,
	userID uuid.UUID,
	userType constant.UserType,
) ([]*entity.Connection, error) {
	query := `
		SELECT id, user_id, connection_id, socket_id,
			user_agent, ip_address, connected_at, last_heartbeat, is_active
		FROM connections
		WHERE user_id = $1 AND is_active = TRUE
		ORDER BY last_heartbeat DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active connections by user: %w", err)
	}
	defer rows.Close()

	return r.scanConnections(rows)
}

// FindAllActive retrieves all active connections.
func (r *connectionRepository) FindAllActive(ctx context.Context) ([]*entity.Connection, error) {
	query := `
		SELECT id, user_id, connection_id, socket_id,
			user_agent, ip_address, connected_at, last_heartbeat, is_active
		FROM connections
		WHERE is_active = TRUE
		ORDER BY last_heartbeat DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all active connections: %w", err)
	}
	defer rows.Close()

	return r.scanConnections(rows)
}

// UpdateHeartbeat updates the last heartbeat timestamp.
func (r *connectionRepository) UpdateHeartbeat(ctx context.Context, connectionID string) error {
	query := `
		UPDATE connections
		SET last_heartbeat = NOW()
		WHERE connection_id = $1 AND is_active = TRUE
	`

	result, err := r.db.Exec(ctx, query, connectionID)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("connection with id %s not found or inactive", connectionID)
	}

	return nil
}

// MarkAsInactive marks a connection as inactive.
func (r *connectionRepository) MarkAsInactive(ctx context.Context, connectionID string) error {
	query := `
		UPDATE connections
		SET is_active = FALSE
		WHERE connection_id = $1
	`

	result, err := r.db.Exec(ctx, query, connectionID)
	if err != nil {
		return fmt.Errorf("failed to mark connection as inactive: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("connection with id %s not found", connectionID)
	}

	return nil
}

// CleanupStaleConnections removes connections that haven't sent heartbeat recently.
func (r *connectionRepository) CleanupStaleConnections(
	ctx context.Context,
	staleThreshold time.Duration,
) (int64, error) {
	query := `
		UPDATE connections
		SET is_active = FALSE
		WHERE is_active = TRUE
		AND last_heartbeat < NOW() - INTERVAL '%d seconds'
	`

	result, err := r.db.Exec(ctx, fmt.Sprintf(query, int(staleThreshold.Seconds())))
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup stale connections: %w", err)
	}

	return result.RowsAffected(), nil
}

// Delete removes a connection.
func (r *connectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM connections WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("connection with id %s not found", id)
	}

	return nil
}

// scanConnection scans a row into a Connection entity.
func (r *connectionRepository) scanConnection(row pgx.Row) (*entity.Connection, error) {
	var connection entity.Connection

	err := row.Scan(
		&connection.ID,
		&connection.UserID,
		&connection.ConnectionID,
		&connection.SocketID,
		&connection.UserAgent,
		&connection.IPAddress,
		&connection.ConnectedAt,
		&connection.LastHeartbeat,
		&connection.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("connection not found")
		}

		return nil, fmt.Errorf("failed to scan connection: %w", err)
	}

	return &connection, nil
}

// scanConnections scans multiple rows into Connection entities.
func (r *connectionRepository) scanConnections(rows pgx.Rows) ([]*entity.Connection, error) {
	var connections []*entity.Connection

	for rows.Next() {
		var connection entity.Connection

		err := rows.Scan(
			&connection.ID,
			&connection.UserID,
			&connection.ConnectionID,
			&connection.SocketID,
			&connection.UserAgent,
			&connection.IPAddress,
			&connection.ConnectedAt,
			&connection.LastHeartbeat,
			&connection.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}

		connections = append(connections, &connection)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return connections, nil
}
