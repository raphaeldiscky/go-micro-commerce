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

// PaymentRepositoryInterface defines the interface for payment data operations.
type PaymentRepositoryInterface interface {
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
}

// PaymentRepositoryPostgres implements the ProductRepository interface for PostgreSQL.
type PaymentRepositoryPostgres struct {
	db DBTX
}

// NewPaymentRepositoryPostgres creates a new instance of PaymentRepositoryPostgres.
func NewPaymentRepositoryPostgres(db DBTX) PaymentRepositoryInterface {
	return &PaymentRepositoryPostgres{
		db: db,
	}
}

// Create creates a new payment.
func (r *PaymentRepositoryPostgres) Create(
	ctx context.Context,
	payment *entity.Payment,
) (*entity.Payment, error) {
	query := `
		INSERT INTO payments (
			id, order_id, amount, currency, status, payment_method, 
			payment_gateway, gateway_reference_id, gateway_response,
			created_at, updated_at, completed_at, failed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, order_id, amount, currency, status, payment_method,
			payment_gateway, gateway_reference_id, gateway_response,
			created_at, updated_at, completed_at, failed_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.PaymentGateway,
		payment.GatewayReferenceID,
		payment.GatewayResponse,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.CompletedAt,
		payment.FailedAt,
	)

	var createdPayment entity.Payment

	err := row.Scan(
		&createdPayment.ID,
		&createdPayment.OrderID,
		&createdPayment.Amount,
		&createdPayment.Currency,
		&createdPayment.Status,
		&createdPayment.PaymentMethod,
		&createdPayment.PaymentGateway,
		&createdPayment.GatewayReferenceID,
		&createdPayment.GatewayResponse,
		&createdPayment.CreatedAt,
		&createdPayment.UpdatedAt,
		&createdPayment.CompletedAt,
		&createdPayment.FailedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return &createdPayment, nil
}

// FindByID retrieves a payment by its ID.
func (r *PaymentRepositoryPostgres) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, status, payment_method,
			payment_gateway, gateway_reference_id, gateway_response,
			created_at, updated_at, completed_at, failed_at
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
		&payment.PaymentMethod,
		&payment.PaymentGateway,
		&payment.GatewayReferenceID,
		&payment.GatewayResponse,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.CompletedAt,
		&payment.FailedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to scan payment: %w", err)
	}

	return &payment, nil
}

// FindByOrderID retrieves a payment by its order ID.
func (r *PaymentRepositoryPostgres) FindByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*entity.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, status, payment_method,
			payment_gateway, gateway_reference_id, gateway_response,
			created_at, updated_at, completed_at, failed_at
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
		&payment.PaymentMethod,
		&payment.PaymentGateway,
		&payment.GatewayReferenceID,
		&payment.GatewayResponse,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.CompletedAt,
		&payment.FailedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to scan payment: %w", err)
	}

	return &payment, nil
}

// Update updates an existing payment.
func (r *PaymentRepositoryPostgres) Update(
	ctx context.Context,
	payment *entity.Payment,
) (*entity.Payment, error) {
	query := `
		UPDATE payments
		SET order_id = $2,
			amount = $3,
			currency = $4,
			status = $5,
			payment_method = $6,
			payment_gateway = $7,
			gateway_reference_id = $8,
			gateway_response = $9,
			updated_at = $10,
			completed_at = $11,
			failed_at = $12
		WHERE id = $1
		RETURNING id, order_id, amount, currency, status, payment_method,
			payment_gateway, gateway_reference_id, gateway_response,
			created_at, updated_at, completed_at, failed_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.PaymentGateway,
		payment.GatewayReferenceID,
		payment.GatewayResponse,
		payment.UpdatedAt,
		payment.CompletedAt,
		payment.FailedAt,
	)

	var updatedPayment entity.Payment

	err := row.Scan(
		&updatedPayment.ID,
		&updatedPayment.OrderID,
		&updatedPayment.Amount,
		&updatedPayment.Currency,
		&updatedPayment.Status,
		&updatedPayment.PaymentMethod,
		&updatedPayment.PaymentGateway,
		&updatedPayment.GatewayReferenceID,
		&updatedPayment.GatewayResponse,
		&updatedPayment.CreatedAt,
		&updatedPayment.UpdatedAt,
		&updatedPayment.CompletedAt,
		&updatedPayment.FailedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // payment not found
		}

		return nil, fmt.Errorf("failed to scan updated payment: %w", err)
	}

	return &updatedPayment, nil
}

// UpdateStatus updates only the status of a payment.
func (r *PaymentRepositoryPostgres) UpdateStatus(
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
