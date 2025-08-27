// Package repository defines the interface for product data operations.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-template/payment-service/internal/entity"
)

// PaymentRepositoryInterface defines the interface for payment data operations.
type PaymentRepositoryInterface interface {
	// Update updates an existing payment
	Update(ctx context.Context, payment *entity.Payment) (*entity.Payment, error)

	// UpdateStatus updates only the status of an payment
	UpdateStatus(ctx context.Context, id uuid.UUID, status constant.PaymentStatus) error

	// FindByID retrieves a payment by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error)
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

// FindByID retrieves a payment by its ID.
func (r *PaymentRepositoryPostgres) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Payment, error) {
	// Get payment
	paymentQuery := `
		SELECT id, idempotency_key, created_at, updated_at, customer_id, status, total_price
		FROM payments
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, paymentQuery, id)

	var payment entity.Payment

	err := row.Scan(
		&payment.ID,
		&payment.IdempotencyKey,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.CustomerID,
		&payment.Status,
		&payment.TotalPrice,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to scan payment: %w", err)
	}

	// Get payment items
	const itemsQuery = `
		SELECT id, payment_id, product_id, quantity, price
		FROM payment_items
		WHERE payment_id = $1
	`

	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment items: %w", err)
	}
	defer rows.Close()

	var items []entity.PaymentItem

	for rows.Next() {
		var item entity.PaymentItem

		err := rows.Scan(
			&item.ID,
			&item.PaymentID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment item: %w", err)
		}

		items = append(items, item)
	}

	payment.Items = items

	return &payment, nil
}

// Update updates an existing payment.
func (r *PaymentRepositoryPostgres) Update(
	ctx context.Context,
	payment *entity.Payment,
) (*entity.Payment, error) {
	// Update the payment itself
	updatePaymentQuery := `
		UPDATE payments
		SET customer_id = $1,
			idempotency_key = $2,
			status = $3,
			total_price = $4,
			updated_at = $5
		WHERE id = $6
		RETURNING id, idempotency_key, customer_id, status, total_price, created_at, updated_at
	`

	row := r.db.QueryRow(
		ctx,
		updatePaymentQuery,
		payment.CustomerID,     // $1
		payment.IdempotencyKey, // $2
		payment.Status,         // $3
		payment.TotalPrice,     // $4
		payment.UpdatedAt,      // $5
		payment.ID,             // $6
	)

	var updatedPayment entity.Payment

	err := row.Scan(
		&updatedPayment.ID,
		&updatedPayment.IdempotencyKey,
		&updatedPayment.CustomerID,
		&updatedPayment.Status,
		&updatedPayment.TotalPrice,
		&updatedPayment.CreatedAt,
		&updatedPayment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // payment not found
		}

		return nil, fmt.Errorf("failed to scan updated payment: %w", err)
	}

	// Delete existing items
	_, err = r.db.Exec(ctx, "DELETE FROM payment_items WHERE payment_id = $1", payment.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing payment items: %w", err)
	}

	// Insert new items if provided
	if len(payment.Items) > 0 {
		insertItemQuery := `
			INSERT INTO payment_items (id, payment_id, product_id, quantity, price)
			VALUES ($1, $2, $3, $4, $5)
		`
		for _, item := range payment.Items {
			_, err = r.db.Exec(
				ctx,
				insertItemQuery,
				item.ID,
				payment.ID,
				item.ProductID,
				item.Quantity,
				item.Price,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to insert payment item: %w", err)
			}
		}
	}

	// Attach updated items back to the result
	updatedPayment.Items = payment.Items

	return &updatedPayment, nil
}

// UpdateStatus updates only the status of an payment.
func (r *PaymentRepositoryPostgres) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status constant.PaymentStatus,
) error {
	query := `
		UPDATE payments
		SET status = $2, updated_at = NOW()
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
