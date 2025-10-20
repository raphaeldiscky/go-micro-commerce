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
			id, order_id, amount, currency, status, payment_method,
			payment_gateway, gateway_reference_id, gateway_response,
			payment_method_id, stripe_customer_id,
			created_at, updated_at, completed_at, failed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, order_id, amount, currency, status, payment_method,
			payment_gateway, gateway_reference_id, gateway_response,
			payment_method_id, stripe_customer_id,
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
		payment.PaymentMethodID,
		payment.StripeCustomerID,
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
		&createdPayment.PaymentMethodID,
		&createdPayment.StripeCustomerID,
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
func (r *paymentRepositoryPostgres) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Payment, error) {
	query := `
		SELECT id, order_id, amount, currency, status, payment_method,
			payment_gateway, gateway_reference_id, gateway_response,
			payment_method_id, stripe_customer_id,
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
		&payment.PaymentMethodID,
		&payment.StripeCustomerID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.CompletedAt,
		&payment.FailedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.PaymentNotFoundErrorMessage)
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
		SELECT id, order_id, amount, currency, status, payment_method,
			payment_gateway, gateway_reference_id, gateway_response,
			payment_method_id, stripe_customer_id,
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
		&payment.PaymentMethodID,
		&payment.StripeCustomerID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.CompletedAt,
		&payment.FailedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.PaymentNotFoundErrorMessage)
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
			payment_method = $6,
			payment_gateway = $7,
			gateway_reference_id = $8,
			gateway_response = $9,
			payment_method_id = $10,
			stripe_customer_id = $11,
			updated_at = $12,
			completed_at = $13,
			failed_at = $14
		WHERE id = $1
		RETURNING id, order_id, amount, currency, status, payment_method,
			payment_gateway, gateway_reference_id, gateway_response,
			payment_method_id, stripe_customer_id,
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
		payment.PaymentMethodID,
		payment.StripeCustomerID,
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
		&updatedPayment.PaymentMethodID,
		&updatedPayment.StripeCustomerID,
		&updatedPayment.CreatedAt,
		&updatedPayment.UpdatedAt,
		&updatedPayment.CompletedAt,
		&updatedPayment.FailedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.PaymentNotFoundErrorMessage)
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
