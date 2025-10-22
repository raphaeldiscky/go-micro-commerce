package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/cart-service/internal/entity"
)

// CheckoutSessionRepository defines the interface for CheckoutSession data operations.
type CheckoutSessionRepository interface {
	// Create saves a new CheckoutSession with its items
	Create(ctx context.Context, session *entity.CheckoutSession) (*entity.CheckoutSession, error)
	// GetByID retrieves a CheckoutSession by its ID with items
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CheckoutSession, error)
	// Update updates an existing CheckoutSession
	Update(
		ctx context.Context,
		checkoutSession *entity.CheckoutSession,
	) (*entity.CheckoutSession, error)
}

// checkoutSessionRepository implements the CheckoutSessionRepository interface for PostgreSQL.
type checkoutSessionRepository struct {
	db DBTX
}

// NewCheckoutSessionRepository creates a new instance of checkoutSessionRepository.
func NewCheckoutSessionRepository(db DBTX) CheckoutSessionRepository {
	return &checkoutSessionRepository{
		db: db,
	}
}

// Create creates a new checkout session in the database.
func (r *checkoutSessionRepository) Create(
	ctx context.Context,
	session *entity.CheckoutSession,
) (*entity.CheckoutSession, error) {
	// Insert checkout session
	insertSessionQuery := `
        INSERT INTO checkout_sessions (
            id, idempotency_key, customer_id, address_id, carrier_id,
            status, payment_gateway, payment_method, currency,
            created_at, updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id, idempotency_key, customer_id, address_id, carrier_id,
                  status, payment_gateway, payment_method, currency,
                  created_at, updated_at
    `

	var createdSession entity.CheckoutSession

	err := r.db.QueryRow(
		ctx,
		insertSessionQuery,
		session.ID,
		session.IdempotencyKey,
		session.CustomerID,
		session.AddressID,
		session.CarrierID,
		session.Status,
		session.PaymentGateway,
		session.PaymentMethod,
		session.Currency,
		session.CreatedAt,
		session.UpdatedAt,
	).Scan(
		&createdSession.ID,
		&createdSession.IdempotencyKey,
		&createdSession.CustomerID,
		&createdSession.AddressID,
		&createdSession.CarrierID,
		&createdSession.Status,
		&createdSession.PaymentGateway,
		&createdSession.PaymentMethod,
		&createdSession.Currency,
		&createdSession.CreatedAt,
		&createdSession.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	// Insert checkout session items
	if len(session.Items) > 0 {
		const insertItemQuery = `
            INSERT INTO checkout_session_items (id, checkout_session_id, product_id, quantity)
            VALUES ($1, $2, $3, $4)
        `

		for i := range len(session.Items) {
			item := &session.Items[i]

			_, err = r.db.Exec(
				ctx,
				insertItemQuery,
				item.ID,
				createdSession.ID,
				item.ProductID,
				item.Quantity,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to insert checkout session item: %w", err)
			}
		}
	}

	createdSession.Items = session.Items

	return &createdSession, nil
}

// GetByID retrieves a checkout session by its ID.
func (r *checkoutSessionRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.CheckoutSession, error) {
	// Get checkout session
	sessionQuery := `
		SELECT id, idempotency_key, customer_id, address_id, carrier_id,
		       status, payment_gateway, payment_method, currency,
		       created_at, updated_at
		FROM checkout_sessions
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, sessionQuery, id)

	var session entity.CheckoutSession

	err := row.Scan(
		&session.ID,
		&session.IdempotencyKey,
		&session.CustomerID,
		&session.AddressID,
		&session.CarrierID,
		&session.Status,
		&session.PaymentGateway,
		&session.PaymentMethod,
		&session.Currency,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("checkout session not found")
		}

		return nil, fmt.Errorf("failed to scan checkout session: %w", err)
	}

	// Get checkout session items (ordered by UUIDv7 id for chronological order)
	const itemsQuery = `
		SELECT id, product_id, quantity
		FROM checkout_session_items
		WHERE checkout_session_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query checkout session items: %w", err)
	}
	defer rows.Close()

	var items []entity.CheckoutSessionItem

	for rows.Next() {
		var item entity.CheckoutSessionItem

		err = rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan checkout session item: %w", err)
		}

		items = append(items, item)
	}

	session.Items = items

	return &session, nil
}

// Update updates an existing checkout session.
func (r *checkoutSessionRepository) Update(
	ctx context.Context,
	session *entity.CheckoutSession,
) (*entity.CheckoutSession, error) {
	updateQuery := `
		UPDATE checkout_sessions
		SET status = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, idempotency_key, customer_id, address_id, carrier_id,
		          status, payment_gateway, payment_method, currency,
		          created_at, updated_at
	`

	var updatedSession entity.CheckoutSession

	err := r.db.QueryRow(
		ctx,
		updateQuery,
		session.Status,
		session.UpdatedAt,
		session.ID,
	).Scan(
		&updatedSession.ID,
		&updatedSession.IdempotencyKey,
		&updatedSession.CustomerID,
		&updatedSession.AddressID,
		&updatedSession.CarrierID,
		&updatedSession.Status,
		&updatedSession.PaymentGateway,
		&updatedSession.PaymentMethod,
		&updatedSession.Currency,
		&updatedSession.UpdatedAt,
		&updatedSession.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.CheckoutSessionNotFoundErrorMessage)
		}

		return nil, fmt.Errorf("failed to update checkout session: %w", err)
	}

	// Get checkout session items (items don't change in update)
	const itemsQuery = `
		SELECT id, product_id, quantity
		FROM checkout_session_items
		WHERE checkout_session_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, itemsQuery, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query checkout session items: %w", err)
	}
	defer rows.Close()

	var items []entity.CheckoutSessionItem

	for rows.Next() {
		var item entity.CheckoutSessionItem

		err = rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan checkout session item: %w", err)
		}

		items = append(items, item)
	}

	updatedSession.Items = items

	return &updatedSession, nil
}
