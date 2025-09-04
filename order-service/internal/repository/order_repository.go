// Package repository defines the interface for product data operations.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// OrderRepositoryInterface defines the interface for order data operations.
type OrderRepositoryInterface interface {
	// Create saves a new order
	Create(ctx context.Context, order *entity.Order) (*entity.Order, error)

	// FindByID retrieves an order by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Order, error)

	// FindByCustomerID retrieves all orders for a specific customer
	FindByCustomerID(
		ctx context.Context,
		customerID uuid.UUID,
		limit, offset int64,
	) ([]*entity.Order, error)

	// FindByIdempotencyKey retrieves an order by its idempotency key
	FindByIdempotencyKey(ctx context.Context, idempotencyKey uuid.UUID) (*entity.Order, error)

	// FindAll retrieves all orders with optional pagination
	FindAll(ctx context.Context, limit, offset int64) ([]*entity.Order, error)

	// Update updates an existing order
	Update(ctx context.Context, order *entity.Order) (*entity.Order, error)

	// Delete removes an order by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateStatus updates only the status of an order
	UpdateStatus(ctx context.Context, id uuid.UUID, status constant.OrderStatus) error

	// Exists checks if an order exists by ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// Count returns the total number of orders
	Count(ctx context.Context) (int64, error)

	// CountByCustomer returns the total number of orders for a specific customer
	CountByCustomer(ctx context.Context, customerID uuid.UUID) (int64, error)
}

// OrderRepositoryPostgres implements the ProductRepository interface for PostgreSQL.
type OrderRepositoryPostgres struct {
	db DBTX
}

// NewOrderRepositoryPostgres creates a new instance of OrderRepositoryPostgres.
func NewOrderRepositoryPostgres(db DBTX) OrderRepositoryInterface {
	return &OrderRepositoryPostgres{
		db: db,
	}
}

// Create creates a new order in the database.
func (r *OrderRepositoryPostgres) Create(
	ctx context.Context,
	order *entity.Order,
) (*entity.Order, error) {
	// Insert order
	insertOrderQuery := `
        INSERT INTO orders (id, idempotency_key, customer_id, status, currency, total_tax, total_discount, total_price, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, idempotency_key, customer_id, status, currency, total_tax, total_discount, total_price, created_at, updated_at
    `

	var createdOrder entity.Order
	// Scan into the existing order object to ensure consistency
	err := r.db.QueryRow(
		ctx,
		insertOrderQuery,
		order.ID,
		order.IdempotencyKey,
		order.CustomerID,
		order.Status,
		order.Currency,
		order.TotalTax,
		order.TotalDiscount,
		order.TotalPrice,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(
		&createdOrder.ID,
		&createdOrder.IdempotencyKey,
		&createdOrder.CustomerID,
		&createdOrder.Status,
		&createdOrder.Currency,
		&createdOrder.TotalTax,
		&createdOrder.TotalDiscount,
		&createdOrder.TotalPrice,
		&createdOrder.CreatedAt,
		&createdOrder.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if len(order.Items) > 0 {
		const insertItemQuery = `
            INSERT INTO order_items (id, order_id, product_id, quantity, currency, unit_price, total_tax, total_discount, total_price, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        `

		for i := 0; i < len(order.Items); i++ {
			item := &order.Items[i]

			_, err = r.db.Exec(
				ctx,
				insertItemQuery,
				item.ID,
				createdOrder.ID,
				item.ProductID,
				item.Quantity,
				item.Currency,
				item.UnitPrice,
				item.TotalTax,
				item.TotalDiscount,
				item.TotalPrice,
				item.CreatedAt,
				item.UpdatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to insert order item: %w", err)
			}
		}
	}

	createdOrder.Items = order.Items

	return &createdOrder, nil
}

// FindByID retrieves an order by its ID.
func (r *OrderRepositoryPostgres) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Order, error) {
	// Get order
	orderQuery := `
		SELECT id, idempotency_key, customer_id, status, currency, total_tax, total_discount, total_price, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, orderQuery, id)

	var order entity.Order

	err := row.Scan(
		&order.ID,
		&order.IdempotencyKey,
		&order.CustomerID,
		&order.Status,
		&order.Currency,
		&order.TotalTax,
		&order.TotalDiscount,
		&order.TotalPrice,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to scan order: %w", err)
	}

	// Get order items
	const itemsQuery = `
		SELECT id, order_id, product_id, quantity, currency, unit_price, total_tax, total_discount, total_price, created_at, updated_at
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []entity.OrderItem

	for rows.Next() {
		var item entity.OrderItem

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Currency,
			&item.UnitPrice,
			&item.TotalTax,
			&item.TotalDiscount,
			&item.TotalPrice,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}

		items = append(items, item)
	}

	order.Items = items

	return &order, nil
}

// FindByIdempotencyKey retrieves an order by its idempotency key.
func (r *OrderRepositoryPostgres) FindByIdempotencyKey(
	ctx context.Context,
	idempotencyKey uuid.UUID,
) (*entity.Order, error) {
	// Get order
	orderQuery := `
		SELECT id, idempotency_key, customer_id, status, currency, total_tax, total_discount, total_price, created_at, updated_at
		FROM orders
		WHERE idempotency_key = $1
	`

	row := r.db.QueryRow(ctx, orderQuery, idempotencyKey)

	var order entity.Order

	err := row.Scan(
		&order.ID,
		&order.IdempotencyKey,
		&order.CustomerID,
		&order.Status,
		&order.Currency,
		&order.TotalTax,
		&order.TotalDiscount,
		&order.TotalPrice,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // no order found
		}

		return nil, fmt.Errorf("failed to scan order: %w", err)
	}

	// Get order items
	const itemsQuery = `
		SELECT id, order_id, product_id, quantity, currency, unit_price, total_tax, total_discount, total_price, created_at, updated_at
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.Query(ctx, itemsQuery, order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []entity.OrderItem

	for rows.Next() {
		var item entity.OrderItem

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Currency,
			&item.UnitPrice,
			&item.TotalTax,
			&item.TotalDiscount,
			&item.TotalPrice,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}

		items = append(items, item)
	}

	order.Items = items

	return &order, nil
}

// FindByCustomerID retrieves all orders for a specific customer.
func (r *OrderRepositoryPostgres) FindByCustomerID(
	ctx context.Context,
	customerID uuid.UUID,
	limit, offset int64,
) ([]*entity.Order, error) {
	query := `
		SELECT id, idempotency_key, customer_id, status, currency, total_tax, total_discount, total_price, created_at, updated_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, customerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*entity.Order

	for rows.Next() {
		var order entity.Order

		err := rows.Scan(
			&order.ID,
			&order.IdempotencyKey,
			&order.CustomerID,
			&order.Status,
			&order.Currency,
			&order.TotalTax,
			&order.TotalDiscount,
			&order.TotalPrice,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		orders = append(orders, &order)
	}

	// Load items for each order
	for _, order := range orders {
		items, err := r.loadOrderItems(ctx, order.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load items for order %s: %w", order.ID, err)
		}

		order.Items = items
	}

	return orders, nil
}

// FindAll retrieves all orders with optional pagination.
func (r *OrderRepositoryPostgres) FindAll(
	ctx context.Context,
	limit, offset int64,
) ([]*entity.Order, error) {
	query := `
		SELECT id, idempotency_key, customer_id, status, currency, total_tax, total_discount, total_price, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*entity.Order

	for rows.Next() {
		var order entity.Order

		err := rows.Scan(
			&order.ID,
			&order.IdempotencyKey,
			&order.CustomerID,
			&order.Status,
			&order.Currency,
			&order.TotalTax,
			&order.TotalDiscount,
			&order.TotalPrice,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		orders = append(orders, &order)
	}

	// Load items for each order
	for _, order := range orders {
		items, err := r.loadOrderItems(ctx, order.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load items for order %s: %w", order.ID, err)
		}

		order.Items = items
	}

	return orders, nil
}

// Update updates an existing order.
func (r *OrderRepositoryPostgres) Update(
	ctx context.Context,
	order *entity.Order,
) (*entity.Order, error) {
	// Update the order itself
	updateOrderQuery := `
		UPDATE orders
		SET customer_id = $1,
			idempotency_key = $2,
			status = $3,
			currency = $4,
			total_tax = $5,
			total_discount = $6,
			total_price = $7,
			updated_at = $8
		WHERE id = $9
		RETURNING id, idempotency_key, customer_id, status, currency, total_tax, total_discount, total_price, created_at, updated_at
	`

	row := r.db.QueryRow(
		ctx,
		updateOrderQuery,
		order.CustomerID,     // $1
		order.IdempotencyKey, // $2
		order.Status,         // $3
		order.Currency,       // $4
		order.TotalTax,       // $5
		order.TotalDiscount,  // $6
		order.TotalPrice,     // $7
		order.UpdatedAt,      // $8
		order.ID,             // $9
	)

	var updatedOrder entity.Order

	err := row.Scan(
		&updatedOrder.ID,
		&updatedOrder.IdempotencyKey,
		&updatedOrder.CustomerID,
		&updatedOrder.Status,
		&updatedOrder.Currency,
		&updatedOrder.TotalTax,
		&updatedOrder.TotalDiscount,
		&updatedOrder.TotalPrice,
		&updatedOrder.CreatedAt,
		&updatedOrder.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // order not found
		}

		return nil, fmt.Errorf("failed to scan updated order: %w", err)
	}

	// Delete existing items
	_, err = r.db.Exec(ctx, "DELETE FROM order_items WHERE order_id = $1", order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing order items: %w", err)
	}

	// Insert new items if provided
	if len(order.Items) > 0 {
		insertItemQuery := `
			INSERT INTO order_items (id, order_id, product_id, quantity, currency, unit_price, total_tax, total_discount, total_price, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`

		for i := 0; i < len(order.Items); i++ {
			item := &order.Items[i]

			_, err = r.db.Exec(
				ctx,
				insertItemQuery,
				item.ID,
				order.ID,
				item.ProductID,
				item.Quantity,
				item.Currency,
				item.UnitPrice,
				item.TotalTax,
				item.TotalDiscount,
				item.TotalPrice,
				item.CreatedAt,
				item.UpdatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to insert order item: %w", err)
			}
		}
	}

	// Attach updated items back to the result
	updatedOrder.Items = order.Items

	return &updatedOrder, nil
}

// Delete removes an order by ID.
func (r *OrderRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete order items first
	_, err := r.db.Exec(ctx, "DELETE FROM order_items WHERE order_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	// Delete order
	result, err := r.db.Exec(ctx, "DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("order with id %s not found", id)
	}

	return nil
}

// UpdateStatus updates only the status of an order.
func (r *OrderRepositoryPostgres) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status constant.OrderStatus,
) error {
	query := `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("order with id %s not found", id)
	}

	return nil
}

// Exists checks if an order exists by ID.
func (r *OrderRepositoryPostgres) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)"

	var exists bool

	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check order existence: %w", err)
	}

	return exists, nil
}

// Count returns the total number of orders.
func (r *OrderRepositoryPostgres) Count(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM orders"

	var count int64

	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}

	return count, nil
}

// CountByCustomer returns the total number of orders for a specific customer.
func (r *OrderRepositoryPostgres) CountByCustomer(
	ctx context.Context,
	customerID uuid.UUID,
) (int64, error) {
	query := "SELECT COUNT(*) FROM orders WHERE customer_id = $1"

	var count int64

	err := r.db.QueryRow(ctx, query, customerID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count customer orders: %w", err)
	}

	return count, nil
}

// loadOrderItems is a helper method to load items for an order.
func (r *OrderRepositoryPostgres) loadOrderItems(
	ctx context.Context,
	orderID uuid.UUID,
) ([]entity.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, currency, unit_price, total_tax, total_discount, total_price, created_at, updated_at
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []entity.OrderItem

	for rows.Next() {
		var item entity.OrderItem

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Currency,
			&item.UnitPrice,
			&item.TotalTax,
			&item.TotalDiscount,
			&item.TotalPrice,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}

		items = append(items, item)
	}

	return items, nil
}
