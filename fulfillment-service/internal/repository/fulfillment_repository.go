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

// FulfillmentRepository defines the interface for fulfillment data operations.
type FulfillmentRepository interface {
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

// fulfillmentRepository implements the FulfillmentRepository interface for PostgreSQL.
type fulfillmentRepository struct {
	db DBTX
}

// NewFulfillmentRepository creates a new instance of fulfillmentRepository.
func NewFulfillmentRepository(db DBTX) FulfillmentRepository {
	return &fulfillmentRepository{
		db: db,
	}
}

// Create creates a new fulfillment.
func (r *fulfillmentRepository) Create(
	ctx context.Context,
	fulfillment *entity.Fulfillment,
) (*entity.Fulfillment, error) {
	var dimensionsJSON []byte

	var err error

	dimensionsJSON, err = json.Marshal(fulfillment.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dimensions: %w", err)
	}

	fromAddressJSON, err := json.Marshal(fulfillment.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal from_address: %w", err)
	}

	toAddressJSON, err := json.Marshal(fulfillment.ToAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to_address: %w", err)
	}

	query := `
		INSERT INTO fulfillments (
			id, order_id, status, tracking_number, carrier_id, 
			shipping_label_url, currency, shipping_cost, from_address, to_address, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, order_id, status, tracking_number, carrier_id,
			shipping_label_url, currency, shipping_cost, from_address, to_address, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		fulfillment.ID,
		fulfillment.OrderID,
		fulfillment.Status,
		fulfillment.TrackingNumber,
		fulfillment.CarrierID,
		fulfillment.ShippingLabelURL,
		fulfillment.Currency,
		fulfillment.ShippingCost,
		fromAddressJSON,
		toAddressJSON,
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
func (r *fulfillmentRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Fulfillment, error) {
	query := `
		SELECT id, order_id, status, tracking_number, carrier_id,
			shipping_label_url, currency, shipping_cost, from_address, to_address, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
		FROM fulfillments
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, query, id)

	return r.scanFulfillment(row)
}

// FindByOrderID retrieves a fulfillment by its order ID.
func (r *fulfillmentRepository) FindByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*entity.Fulfillment, error) {
	query := `
		SELECT id, order_id, status, tracking_number, carrier_id,
			shipping_label_url, currency, shipping_cost, from_address, to_address, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
		FROM fulfillments
		WHERE order_id = $1
	`

	row := r.db.QueryRow(ctx, query, orderID)

	return r.scanFulfillment(row)
}

// FindByTrackingNumber retrieves a fulfillment by its tracking number.
func (r *fulfillmentRepository) FindByTrackingNumber(
	ctx context.Context,
	trackingNumber string,
) (*entity.Fulfillment, error) {
	query := `
		SELECT id, order_id, status, tracking_number, carrier_id,
			shipping_label_url, currency, shipping_cost, from_address, to_address, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
		FROM fulfillments
		WHERE tracking_number = $1
	`

	row := r.db.QueryRow(ctx, query, trackingNumber)

	return r.scanFulfillment(row)
}

// Update updates an existing fulfillment.
func (r *fulfillmentRepository) Update(
	ctx context.Context,
	fulfillment *entity.Fulfillment,
) (*entity.Fulfillment, error) {
	var dimensionsJSON []byte

	var err error

	dimensionsJSON, err = json.Marshal(fulfillment.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dimensions: %w", err)
	}

	fromAddressJSON, err := json.Marshal(fulfillment.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal from_address: %w", err)
	}

	toAddressJSON, err := json.Marshal(fulfillment.ToAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to_address: %w", err)
	}

	query := `
		UPDATE fulfillments
		SET order_id = $2,
			status = $3,
			tracking_number = $4,
			carrier_id = $5,
			shipping_label_url = $6,
			currency = $7,
			shipping_cost = $8,
			from_address = $9,
			to_address = $10,
			weight_kg = $11,
			dimensions = $12,
			estimated_delivery_at = $13,
			actual_delivery_at = $14,
			updated_at = $15
		WHERE id = $1
		RETURNING id, order_id, status, tracking_number, carrier_id,
			shipping_label_url, currency, shipping_cost, from_address, to_address, weight_kg, dimensions,
			estimated_delivery_at, actual_delivery_at, created_at, updated_at
	`

	row := r.db.QueryRow(
		ctx,
		query,
		fulfillment.ID,
		fulfillment.OrderID,
		fulfillment.Status,
		fulfillment.TrackingNumber,
		fulfillment.CarrierID,
		fulfillment.ShippingLabelURL,
		fulfillment.Currency,
		fulfillment.ShippingCost,
		fromAddressJSON,
		toAddressJSON,
		fulfillment.WeightKG,
		dimensionsJSON,
		fulfillment.EstimatedDeliveryAt,
		fulfillment.ActualDeliveryAt,
		fulfillment.UpdatedAt,
	)

	return r.scanFulfillment(row)
}

// UpdateStatus updates only the status of a fulfillment.
func (r *fulfillmentRepository) UpdateStatus(
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
func (r *fulfillmentRepository) scanFulfillment(row pgx.Row) (*entity.Fulfillment, error) {
	var fulfillment entity.Fulfillment

	var dimensionsJSON []byte

	var fromAddressJSON []byte

	var toAddressJSON []byte

	err := row.Scan(
		&fulfillment.ID,
		&fulfillment.OrderID,
		&fulfillment.Status,
		&fulfillment.TrackingNumber,
		&fulfillment.CarrierID,
		&fulfillment.ShippingLabelURL,
		&fulfillment.Currency,
		&fulfillment.ShippingCost,
		&fromAddressJSON,
		&toAddressJSON,
		&fulfillment.WeightKG,
		&dimensionsJSON,
		&fulfillment.EstimatedDeliveryAt,
		&fulfillment.ActualDeliveryAt,
		&fulfillment.CreatedAt,
		&fulfillment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.FulfillmentNotFoundErrorMessage)
		}

		return nil, fmt.Errorf("failed to scan fulfillment: %w", err)
	}

	// Unmarshal from_address
	if err = json.Unmarshal(fromAddressJSON, &fulfillment.FromAddress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal from_address: %w", err)
	}

	// Unmarshal to_address
	if err = json.Unmarshal(toAddressJSON, &fulfillment.ToAddress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to_address: %w", err)
	}

	// Unmarshal dimensions if present
	if dimensionsJSON != nil {
		var dimensions entity.Dimensions
		if err = json.Unmarshal(dimensionsJSON, &dimensions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal dimensions: %w", err)
		}

		fulfillment.Dimensions = dimensions
	}

	return &fulfillment, nil
}
