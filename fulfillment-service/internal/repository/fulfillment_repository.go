// Package repository defines the interface for fulfillment data operations.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/fulfillment-service/internal/entity"
)

// FulfillmentRepositoryInterface defines the interface for fulfillment data operations.
type FulfillmentRepositoryInterface interface {
	// Create creates a new fulfillment
	Create(ctx context.Context, fulfillment *entity.Fulfillment) (*entity.Fulfillment, error)

	// Update updates an existing fulfillment
	Update(ctx context.Context, fulfillment *entity.Fulfillment) (*entity.Fulfillment, error)

	// UpdateStatus updates only the status of a fulfillment
	UpdateStatus(ctx context.Context, id uuid.UUID, status constant.FulfillmentStatus) error

	// FindByID retrieves a fulfillment by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Fulfillment, error)

	// FindByOrderID retrieves a fulfillment by its order ID.
	FindByOrderID(ctx context.Context, orderID uuid.UUID) (*entity.Fulfillment, error)

	// FindByTrackingNumber retrieves a fulfillment by its tracking number.
	FindByTrackingNumber(ctx context.Context, trackingNumber string) (*entity.Fulfillment, error)
}

// FulfillmentRepositoryPostgres implements the FulfillmentRepository interface for PostgreSQL.
type FulfillmentRepositoryPostgres struct {
	db DBTX
}

// NewFulfillmentRepositoryPostgres creates a new instance of FulfillmentRepositoryPostgres.
func NewFulfillmentRepositoryPostgres(db DBTX) FulfillmentRepositoryInterface {
	return &FulfillmentRepositoryPostgres{
		db: db,
	}
}

// Create creates a new fulfillment.
func (r *FulfillmentRepositoryPostgres) Create(
	ctx context.Context,
	fulfillment *entity.Fulfillment,
) (*entity.Fulfillment, error) {
	var dimensionsJSON []byte

	var err error
	if fulfillment.Dimensions != nil {
		dimensionsJSON, err = json.Marshal(fulfillment.Dimensions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal dimensions: %w", err)
		}
	}

	query := `
		INSERT INTO fulfillments (
			id, order_id, status, tracking_number, carrier, 
			shipping_label_url, currency, shipping_cost, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, order_id, status, tracking_number, carrier,
			shipping_label_url, currency, shipping_cost, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		fulfillment.ID,
		fulfillment.OrderID,
		fulfillment.Status,
		fulfillment.TrackingNumber,
		fulfillment.Carrier,
		fulfillment.ShippingLabelURL,
		fulfillment.Currency,
		fulfillment.ShippingCost,
		fulfillment.WeightKG,
		dimensionsJSON,
		fulfillment.EstimatedDeliveryAt,
		fulfillment.ActualDeliveryAt,
		fulfillment.CreatedAt,
		fulfillment.UpdatedAt,
	)

	return r.scanFulfillment(row)
}

// FindByID retrieves a fulfillment by its ID.
func (r *FulfillmentRepositoryPostgres) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Fulfillment, error) {
	query := `
		SELECT id, order_id, status, tracking_number, carrier,
			shipping_label_url, currency, shipping_cost, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
		FROM fulfillments
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	return r.scanFulfillment(row)
}

// FindByOrderID retrieves a fulfillment by its order ID.
func (r *FulfillmentRepositoryPostgres) FindByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*entity.Fulfillment, error) {
	query := `
		SELECT id, order_id, status, tracking_number, carrier,
			shipping_label_url, currency, shipping_cost, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
		FROM fulfillments
		WHERE order_id = $1
	`

	row := r.db.QueryRow(ctx, query, orderID)

	return r.scanFulfillment(row)
}

// FindByTrackingNumber retrieves a fulfillment by its tracking number.
func (r *FulfillmentRepositoryPostgres) FindByTrackingNumber(
	ctx context.Context,
	trackingNumber string,
) (*entity.Fulfillment, error) {
	query := `
		SELECT id, order_id, status, tracking_number, carrier,
			shipping_label_url, currency, shipping_cost, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
		FROM fulfillments
		WHERE tracking_number = $1
	`

	row := r.db.QueryRow(ctx, query, trackingNumber)

	return r.scanFulfillment(row)
}

// Update updates an existing fulfillment.
func (r *FulfillmentRepositoryPostgres) Update(
	ctx context.Context,
	fulfillment *entity.Fulfillment,
) (*entity.Fulfillment, error) {
	var dimensionsJSON []byte

	var err error
	if fulfillment.Dimensions != nil {
		dimensionsJSON, err = json.Marshal(fulfillment.Dimensions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal dimensions: %w", err)
		}
	}

	query := `
		UPDATE fulfillments
		SET order_id = $2,
			status = $3,
			tracking_number = $4,
			carrier = $5,
			shipping_label_url = $6,
			currency = $7,
			shipping_cost = $8,
			weight_kg = $9,
			dimensions = $10,
			estimated_delivery_at = $11,
			actual_delivery_at = $12,
			updated_at = $13
		WHERE id = $1
		RETURNING id, order_id, status, tracking_number, carrier,
			shipping_label_url, currency, shipping_cost, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		fulfillment.ID,
		fulfillment.OrderID,
		fulfillment.Status,
		fulfillment.TrackingNumber,
		fulfillment.Carrier,
		fulfillment.ShippingLabelURL,
		fulfillment.Currency,
		fulfillment.ShippingCost,
		fulfillment.WeightKG,
		dimensionsJSON,
		fulfillment.EstimatedDeliveryAt,
		fulfillment.ActualDeliveryAt,
		fulfillment.UpdatedAt,
	)

	return r.scanFulfillment(row)
}

// UpdateStatus updates only the status of a fulfillment.
func (r *FulfillmentRepositoryPostgres) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status constant.FulfillmentStatus,
) error {
	query := `
		UPDATE fulfillments
		SET status = $2, 
			updated_at = NOW(),
			actual_delivery_at = CASE WHEN $2 = 'delivered' AND actual_delivery_at IS NULL THEN NOW() ELSE actual_delivery_at END
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update fulfillment status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("fulfillment with id %s not found", id)
	}

	return nil
}

// scanFulfillment scans a row into a Fulfillment entity.
func (r *FulfillmentRepositoryPostgres) scanFulfillment(row pgx.Row) (*entity.Fulfillment, error) {
	var fulfillment entity.Fulfillment

	var dimensionsJSON []byte

	err := row.Scan(
		&fulfillment.ID,
		&fulfillment.OrderID,
		&fulfillment.Status,
		&fulfillment.TrackingNumber,
		&fulfillment.Carrier,
		&fulfillment.ShippingLabelURL,
		&fulfillment.Currency,
		&fulfillment.ShippingCost,
		&fulfillment.WeightKG,
		&dimensionsJSON,
		&fulfillment.EstimatedDeliveryAt,
		&fulfillment.ActualDeliveryAt,
		&fulfillment.CreatedAt,
		&fulfillment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to scan fulfillment: %w", err)
	}

	// Unmarshal dimensions if present
	if dimensionsJSON != nil {
		var dimensions entity.Dimensions
		if err := json.Unmarshal(dimensionsJSON, &dimensions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dimensions: %w", err)
		}

		fulfillment.Dimensions = &dimensions
	}

	return &fulfillment, nil
}
