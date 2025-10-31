package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/telemetry"

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
	db  DBTX
	tel *telemetry.Telemetry
}

// NewCheckoutSessionRepository creates a new instance of checkoutSessionRepository.
func NewCheckoutSessionRepository(db DBTX, tel *telemetry.Telemetry) CheckoutSessionRepository {
	return &checkoutSessionRepository{
		db:  db,
		tel: tel,
	}
}

// Create creates a new checkout session in the database.
func (r *checkoutSessionRepository) Create(
	ctx context.Context,
	session *entity.CheckoutSession,
) (*entity.CheckoutSession, error) {
	ctx, end := r.tel.StartSpan(ctx, "CheckoutSessionRepository.Create")
	defer end()

	r.tel.AddSpanAttributes(ctx, map[string]any{
		"db.operation":    "insert",
		"db.table":        "checkout_sessions",
		"session.id":      session.ID.String(),
		"customer.id":     session.CustomerID.String(),
		"cart.id":         session.CartID.String(),
		"idempotency.key": session.IdempotencyKey,
		"items.count":     len(session.Items),
		"payment.gateway": session.PaymentGateway,
	})

	// Marshal JSONB fields
	courierJSON, err := sonic.Marshal(session.Courier)
	if err != nil {
		r.tel.SetSpanError(ctx, err)
		return nil, fmt.Errorf("failed to marshal courier: %w", err)
	}

	destinationJSON, err := sonic.Marshal(session.Destination)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal destination: %w", err)
	}

	originJSON, err := sonic.Marshal(session.Origin)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal origin: %w", err)
	}

	packageJSON, err := sonic.Marshal(session.Package)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package: %w", err)
	}

	// Insert checkout session
	insertSessionQuery := `
        INSERT INTO checkout_sessions (
            id, idempotency_key, customer_id, cart_id, courier, destination, origin, package,
            status, payment_gateway, currency,
            created_at, updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        RETURNING id, idempotency_key, customer_id, cart_id, courier, destination, origin, package,
                  status, payment_gateway, currency,
                  created_at, updated_at
    `

	var (
		createdSession                                        entity.CheckoutSession
		courierData, destinationData, originData, packageData []byte
	)

	err = r.db.QueryRow(
		ctx,
		insertSessionQuery,
		session.ID,
		session.IdempotencyKey,
		session.CustomerID,
		session.CartID,
		courierJSON,
		destinationJSON,
		originJSON,
		packageJSON,
		session.Status,
		session.PaymentGateway,
		session.Currency,
		session.CreatedAt,
		session.UpdatedAt,
	).Scan(
		&createdSession.ID,
		&createdSession.IdempotencyKey,
		&createdSession.CustomerID,
		&createdSession.CartID,
		&courierData,
		&destinationData,
		&originData,
		&packageData,
		&createdSession.Status,
		&createdSession.PaymentGateway,
		&createdSession.Currency,
		&createdSession.CreatedAt,
		&createdSession.UpdatedAt,
	)
	if err != nil {
		r.tel.SetSpanError(ctx, err)
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	// Unmarshal JSONB fields
	if err = sonic.Unmarshal(courierData, &createdSession.Courier); err != nil {
		return nil, fmt.Errorf("failed to unmarshal courier: %w", err)
	}

	if err = sonic.Unmarshal(destinationData, &createdSession.Destination); err != nil {
		return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
	}

	if err = sonic.Unmarshal(originData, &createdSession.Origin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
	}

	if err = sonic.Unmarshal(packageData, &createdSession.Package); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}

	// Insert checkout session items
	if len(session.Items) > 0 {
		const insertItemQuery = `
            INSERT INTO checkout_session_items (id, checkout_session_id, product_id, product_name, quantity, unit_price)
            VALUES ($1, $2, $3, $4, $5, $6)
        `

		for i := range len(session.Items) {
			item := &session.Items[i]

			_, err = r.db.Exec(
				ctx,
				insertItemQuery,
				item.ID,
				createdSession.ID,
				item.ProductID,
				item.ProductName,
				item.Quantity,
				item.UnitPrice,
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
	ctx, end := r.tel.StartSpan(ctx, "CheckoutSessionRepository.GetByID")
	defer end()

	r.tel.AddSpanAttributes(ctx, map[string]any{
		"db.operation": "select",
		"db.table":     "checkout_sessions",
		"session.id":   id.String(),
	})

	// Get checkout session
	sessionQuery := `
		SELECT id, idempotency_key, customer_id, cart_id, courier, destination, origin, package,
		       status, payment_gateway, currency,
		       created_at, updated_at
		FROM checkout_sessions
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, sessionQuery, id)

	var (
		session                                               entity.CheckoutSession
		courierData, destinationData, originData, packageData []byte
	)

	err := row.Scan(
		&session.ID,
		&session.IdempotencyKey,
		&session.CustomerID,
		&session.CartID,
		&courierData,
		&destinationData,
		&originData,
		&packageData,
		&session.Status,
		&session.PaymentGateway,
		&session.Currency,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.tel.AddSpanAttributes(ctx, map[string]any{
				"session.found": false,
			})

			return nil, errors.New("checkout session not found")
		}

		r.tel.SetSpanError(ctx, err)

		return nil, fmt.Errorf("failed to scan checkout session: %w", err)
	}

	r.tel.AddSpanAttributes(ctx, map[string]any{
		"session.found": true,
		"customer.id":   session.CustomerID.String(),
		"cart.id":       session.CartID.String(),
	})

	// Unmarshal JSONB fields
	if err = sonic.Unmarshal(courierData, &session.Courier); err != nil {
		return nil, fmt.Errorf("failed to unmarshal courier: %w", err)
	}

	if err = sonic.Unmarshal(destinationData, &session.Destination); err != nil {
		return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
	}

	if err = sonic.Unmarshal(originData, &session.Origin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
	}

	if err = sonic.Unmarshal(packageData, &session.Package); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}

	// Get checkout session items
	const itemsQuery = `
		SELECT id, product_id, product_name, quantity, unit_price
		FROM checkout_session_items
		WHERE checkout_session_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		r.tel.SetSpanError(ctx, err)
		return nil, fmt.Errorf("failed to query checkout session items: %w", err)
	}
	defer rows.Close()

	var items []entity.CheckoutSessionItem

	for rows.Next() {
		var item entity.CheckoutSessionItem

		err = rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
			&item.UnitPrice,
		)
		if err != nil {
			r.tel.SetSpanError(ctx, err)
			return nil, fmt.Errorf("failed to scan checkout session item: %w", err)
		}

		items = append(items, item)
	}

	session.Items = items

	r.tel.AddSpanAttributes(ctx, map[string]any{
		"items.count": len(items),
	})

	return &session, nil
}

// Update updates an existing checkout session.
func (r *checkoutSessionRepository) Update(
	ctx context.Context,
	session *entity.CheckoutSession,
) (*entity.CheckoutSession, error) {
	ctx, end := r.tel.StartSpan(ctx, "CheckoutSessionRepository.Update")
	defer end()

	r.tel.AddSpanAttributes(ctx, map[string]any{
		"db.operation":    "update",
		"db.table":        "checkout_sessions",
		"session.id":      session.ID.String(),
		"payment.gateway": session.PaymentGateway,
		"session.status":  string(session.Status),
	})

	// Marshal JSONB fields
	courierJSON, err := sonic.Marshal(session.Courier)
	if err != nil {
		r.tel.SetSpanError(ctx, err)
		return nil, fmt.Errorf("failed to marshal courier: %w", err)
	}

	destinationJSON, err := sonic.Marshal(session.Destination)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal destination: %w", err)
	}

	originJSON, err := sonic.Marshal(session.Origin)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal origin: %w", err)
	}

	packageJSON, err := sonic.Marshal(session.Package)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package: %w", err)
	}

	updateQuery := `
		UPDATE checkout_sessions
		SET courier = $1, destination = $2, origin = $3, package = $4, payment_gateway = $5,
		    status = $6, updated_at = $7
		WHERE id = $8
		RETURNING id, idempotency_key, customer_id, cart_id, courier, destination, origin, package,
		          status, payment_gateway, currency,
		          created_at, updated_at
	`

	var (
		updatedSession                                        entity.CheckoutSession
		courierData, destinationData, originData, packageData []byte
	)

	err = r.db.QueryRow(
		ctx,
		updateQuery,
		courierJSON,
		destinationJSON,
		originJSON,
		packageJSON,
		session.PaymentGateway,
		session.Status,
		session.UpdatedAt,
		session.ID,
	).Scan(
		&updatedSession.ID,
		&updatedSession.IdempotencyKey,
		&updatedSession.CustomerID,
		&updatedSession.CartID,
		&courierData,
		&destinationData,
		&originData,
		&packageData,
		&updatedSession.Status,
		&updatedSession.PaymentGateway,
		&updatedSession.Currency,
		&updatedSession.CreatedAt,
		&updatedSession.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.tel.AddSpanAttributes(ctx, map[string]any{
				"session.found": false,
			})

			return nil, errors.New(constant.CheckoutSessionNotFoundErrorMessage)
		}

		r.tel.SetSpanError(ctx, err)

		return nil, fmt.Errorf("failed to update checkout session: %w", err)
	}

	// Unmarshal JSONB fields
	if err = sonic.Unmarshal(courierData, &updatedSession.Courier); err != nil {
		return nil, fmt.Errorf("failed to unmarshal courier: %w", err)
	}

	if err = sonic.Unmarshal(destinationData, &updatedSession.Destination); err != nil {
		return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
	}

	if err = sonic.Unmarshal(originData, &updatedSession.Origin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
	}

	if err = sonic.Unmarshal(packageData, &updatedSession.Package); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}

	// Get checkout session items (items don't change in update)
	const itemsQuery = `
		SELECT id, product_id, product_name, quantity, unit_price
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
			&item.ProductName,
			&item.Quantity,
			&item.UnitPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan checkout session item: %w", err)
		}

		items = append(items, item)
	}

	updatedSession.Items = items

	return &updatedSession, nil
}
