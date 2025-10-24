// Package repository defines the interface for product data operations.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/payment-service/internal/entity"
)

// PaymentRepository defines the interface for payment data operations.
type PaymentRepository interface {
	// Create creates a new payment
	Create(ctx context.Context, payment *entity.Payment) (*entity.Payment, error)

	// Update updates an existing payment
	Update(ctx context.Context, payment *entity.Payment) (*entity.Payment, error)

	// UpdateStatus updates only the status of an payment
	UpdateStatus(ctx context.Context, id uuid.UUID, status constant.PaymentStatus) error

	// FindByID retrieves a payment by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error)

	// FindByOrderID retrieves a payment by its order ID.
	FindByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.Payment, error)

	// FindExpiredPayments finds all pending payments that have expired
	FindExpiredPayments(ctx context.Context, limit int) ([]*entity.Payment, error)
}

// paymentRepositoryPostgres implements the ProductRepository interface for PostgreSQL.
type paymentRepositoryPostgres struct {
	db DBTX
}

// NewPaymentRepository creates a new instance of paymentRepositoryPostgres.
func NewPaymentRepository(db DBTX) PaymentRepository {
	return &paymentRepositoryPostgres{
		db: db,
	}
}

// Create creates a new payment.
func (r *paymentRepositoryPostgres) Create(
	ctx context.Context,
	payment *entity.Payment,
) (*entity.Payment, error) {
	query := `
		INSERT INTO payments (
			id, order_id, amount, currency, status,
			payment_gateway, gateway_transaction_id, gateway_metadata,
			created_at, updated_at, completed_at, failed_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, order_id, amount, currency, status,
			payment_gateway, gateway_transaction_id, gateway_metadata,
			created_at, updated_at, completed_at, failed_at, expires_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentGateway,
		payment.GatewayTransactionID,
		payment.GatewayMetadata,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.CompletedAt,
		payment.FailedAt,
		payment.ExpiresAt,
	)

	var createdPayment entity.Payment

	err := row.Scan(
		&createdPayment.ID,
		&createdPayment.OrderID,
		&createdPayment.Amount,
		&createdPayment.Currency,
		&createdPayment.Status,
		&createdPayment.PaymentGateway,
		&createdPayment.GatewayTransactionID,
		&createdPayment.GatewayMetadata,
		&createdPayment.CreatedAt,
		&createdPayment.UpdatedAt,
		&createdPayment.CompletedAt,
		&createdPayment.FailedAt,
		&createdPayment.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return &createdPayment, nil
}

// FindByID retrieves a payment by its ID.
func (r *paymentRepositoryPostgres) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, status,
			payment_gateway, gateway_transaction_id, gateway_metadata,
			created_at, updated_at, completed_at, failed_at, expires_at
		FROM payments
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	var payment entity.Payment

	err := row.Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentGateway,
		&payment.GatewayTransactionID,
		&payment.GatewayMetadata,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.CompletedAt,
		&payment.FailedAt,
		&payment.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, constant.ErrPaymentNotFound
		}

		return nil, fmt.Errorf("failed to scan payment: %w", err)
	}

	return &payment, nil
}

// FindByOrderID retrieves a payment by its order ID.
func (r *paymentRepositoryPostgres) FindByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*entity.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, status,
			payment_gateway, gateway_transaction_id, gateway_metadata,
			created_at, updated_at, completed_at, failed_at, expires_at
		FROM payments
		WHERE order_id = $1
	`

	row := r.db.QueryRow(ctx, query, orderID)

	var payment entity.Payment

	err := row.Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentGateway,
		&payment.GatewayTransactionID,
		&payment.GatewayMetadata,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.CompletedAt,
		&payment.FailedAt,
		&payment.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, constant.ErrPaymentNotFound
		}

		return nil, fmt.Errorf("failed to scan payment: %w", err)
	}

	return &payment, nil
}

// Update updates an existing payment.
func (r *paymentRepositoryPostgres) Update(
	ctx context.Context,
	payment *entity.Payment,
) (*entity.Payment, error) {
	query := `
		UPDATE payments
		SET order_id = $2,
			amount = $3,
			currency = $4,
			status = $5,
			payment_gateway = $6,
			gateway_transaction_id = $7,
			gateway_metadata = $8,
			updated_at = $9,
			completed_at = $10,
			failed_at = $11,
			expires_at = $12
		WHERE id = $1
		RETURNING id, order_id, amount, currency, status,
			payment_gateway, gateway_transaction_id, gateway_metadata,
			created_at, updated_at, completed_at, failed_at, expires_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentGateway,
		payment.GatewayTransactionID,
		payment.GatewayMetadata,
		payment.UpdatedAt,
		payment.CompletedAt,
		payment.FailedAt,
		payment.ExpiresAt,
	)

	var updatedPayment entity.Payment

	err := row.Scan(
		&updatedPayment.ID,
		&updatedPayment.OrderID,
		&updatedPayment.Amount,
		&updatedPayment.Currency,
		&updatedPayment.Status,
		&updatedPayment.PaymentGateway,
		&updatedPayment.GatewayTransactionID,
		&updatedPayment.GatewayMetadata,
		&updatedPayment.CreatedAt,
		&updatedPayment.UpdatedAt,
		&updatedPayment.CompletedAt,
		&updatedPayment.FailedAt,
		&updatedPayment.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, constant.ErrPaymentNotFound
		}

		return nil, fmt.Errorf("failed to scan updated payment: %w", err)
	}

	return &updatedPayment, nil
}

// UpdateStatus updates only the status of a payment.
func (r *paymentRepositoryPostgres) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status constant.PaymentStatus,
) error {
	query := `
		UPDATE payments
		SET status = $2,
			updated_at = NOW(),
			completed_at = CASE WHEN $2 = 'completed' THEN NOW() ELSE completed_at END,
			failed_at = CASE WHEN $2 = 'failed' THEN NOW() ELSE failed_at END
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("payment with id %s not found", id)
	}

	return nil
}

// FindExpiredPayments finds all pending payments that have expired (past expires_at timestamp).
// Used by the payment timeout job to automatically timeout expired payments.
func (r *paymentRepositoryPostgres) FindExpiredPayments(
	ctx context.Context,
	limit int,
) ([]*entity.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, status,
			payment_gateway, gateway_transaction_id, gateway_metadata,
			created_at, updated_at, completed_at, failed_at, expires_at
		FROM payments
		WHERE status = 'pending'
			AND expires_at IS NOT NULL
			AND expires_at < NOW()
		ORDER BY expires_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired payments: %w", err)
	}
	defer rows.Close()

	payments := make([]*entity.Payment, 0)

	for rows.Next() {
		var payment entity.Payment

		scanErr := rows.Scan(
			&payment.ID,
			&payment.OrderID,
			&payment.Amount,
			&payment.Currency,
			&payment.Status,
			&payment.PaymentGateway,
			&payment.GatewayTransactionID,
			&payment.GatewayMetadata,
			&payment.CreatedAt,
			&payment.UpdatedAt,
			&payment.CompletedAt,
			&payment.FailedAt,
			&payment.ExpiresAt,
		)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", scanErr)
		}

		payments = append(payments, &payment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payments: %w", err)
	}

	return payments, nil
}
