// Package repository defines the interface for product data operations.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/constant"
	"github.com/raphaeldiscky/go-micro-commerce/order-service/internal/entity"
)

// OrderRepository defines the interface for order data operations.
type OrderRepository interface {
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

	// FindByCustomerIDWithCursor retrieves orders for a customer with cursor-based pagination
	FindByCustomerIDWithCursor(
		ctx context.Context,
		customerID uuid.UUID,
		limit int64,
		cursorID string,
		cursorTimestamp int64,
	) ([]*entity.Order, error)

	// FindByIdempotencyKey retrieves an order by its idempotency key
	FindByIdempotencyKey(ctx context.Context, idempotencyKey uuid.UUID) (*entity.Order, error)

	// FindAll retrieves all orders with optional pagination
	FindAll(ctx context.Context, limit, offset int64) ([]*entity.Order, error)

	// FindAllWithCursor retrieves all orders with cursor-based pagination
	FindAllWithCursor(
		ctx context.Context,
		limit int64,
		cursorID string,
		cursorTimestamp int64,
	) ([]*entity.Order, error)

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

// orderRepository implements the ProductRepository interface for PostgreSQL.
type orderRepository struct {
	db DBTX
}

// NewOrderRepository creates a new instance of orderRepository.
func NewOrderRepository(db DBTX) OrderRepository {
	return &orderRepository{
		db: db,
	}
}

// Create creates a new order in the database.
func (r *orderRepository) Create(
	ctx context.Context,
	order *entity.Order,
) (*entity.Order, error) {
	// Marshal JSONB fields
	courierJSON, err := sonic.Marshal(order.Courier)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal courier: %w", err)
	}

	destinationJSON, err := sonic.Marshal(order.Destination)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal destination: %w", err)
	}

	originJSON, err := sonic.Marshal(order.Origin)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal origin: %w", err)
	}

	packageJSON, err := sonic.Marshal(order.Package)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package: %w", err)
	}

	// Insert order
	insertOrderQuery := `
        INSERT INTO orders (id, idempotency_key, checkout_session_id, customer_id, status, payment_gateway, currency, courier, destination, origin, package, shipping_cost, subtotal, total_tax, total_discount, total_price, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
        RETURNING id, idempotency_key, checkout_session_id, customer_id, status, payment_gateway, currency, courier, destination, origin, package, shipping_cost, subtotal, total_tax, total_discount, total_price, created_at, updated_at
    `

	var (
		createdOrder                                          entity.Order
		courierData, destinationData, originData, packageData []byte
	)

	err = r.db.QueryRow(
		ctx,
		insertOrderQuery,
		order.ID,
		order.IdempotencyKey,
		order.CheckoutSessionID,
		order.CustomerID,
		order.Status,
		order.PaymentGateway,
		order.Currency,
		courierJSON,
		destinationJSON,
		originJSON,
		packageJSON,
		order.ShippingCost,
		order.Subtotal,
		order.TotalTax,
		order.TotalDiscount,
		order.TotalPrice,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(
		&createdOrder.ID,
		&createdOrder.IdempotencyKey,
		&createdOrder.CheckoutSessionID,
		&createdOrder.CustomerID,
		&createdOrder.Status,
		&createdOrder.PaymentGateway,
		&createdOrder.Currency,
		&courierData,
		&destinationData,
		&originData,
		&packageData,
		&createdOrder.ShippingCost,
		&createdOrder.Subtotal,
		&createdOrder.TotalTax,
		&createdOrder.TotalDiscount,
		&createdOrder.TotalPrice,
		&createdOrder.CreatedAt,
		&createdOrder.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Unmarshal JSONB fields
	if err = sonic.Unmarshal(courierData, &createdOrder.Courier); err != nil {
		return nil, fmt.Errorf("failed to unmarshal courier: %w", err)
	}

	if err = sonic.Unmarshal(destinationData, &createdOrder.Destination); err != nil {
		return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
	}

	if err = sonic.Unmarshal(originData, &createdOrder.Origin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
	}

	if err = sonic.Unmarshal(packageData, &createdOrder.Package); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}

	if len(order.Items) > 0 {
		const insertItemQuery = `
            INSERT INTO order_items (id, order_id, product_id, product_name, quantity, tax_rate, unit_price, total_tax, total_discount, total_price, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        `

		for i := range len(order.Items) {
			item := &order.Items[i]

			_, err = r.db.Exec(
				ctx,
				insertItemQuery,
				item.ID,
				createdOrder.ID,
				item.ProductID,
				item.ProductName,
				item.Quantity,
				item.TaxRate,
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
func (r *orderRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.Order, error) {
	// Get order
	orderQuery := `
		SELECT id, idempotency_key, checkout_session_id, customer_id, status, payment_gateway, currency, courier, destination, origin, package, shipping_cost, subtotal, total_tax, total_discount, total_price, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, orderQuery, id)

	var (
		order                                                 entity.Order
		courierData, destinationData, originData, packageData []byte
	)

	err := row.Scan(
		&order.ID,
		&order.IdempotencyKey,
		&order.CheckoutSessionID,
		&order.CustomerID,
		&order.Status,
		&order.PaymentGateway,
		&order.Currency,
		&courierData,
		&destinationData,
		&originData,
		&packageData,
		&order.ShippingCost,
		&order.Subtotal,
		&order.TotalTax,
		&order.TotalDiscount,
		&order.TotalPrice,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.OrderNotFoundErrorMessage)
		}

		return nil, fmt.Errorf("failed to scan order: %w", err)
	}

	// Unmarshal JSONB fields
	if err = sonic.Unmarshal(courierData, &order.Courier); err != nil {
		return nil, fmt.Errorf("failed to unmarshal courier: %w", err)
	}

	if err = sonic.Unmarshal(destinationData, &order.Destination); err != nil {
		return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
	}

	if err = sonic.Unmarshal(originData, &order.Origin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
	}

	if err = sonic.Unmarshal(packageData, &order.Package); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}

	// Get order items
	const itemsQuery = `
		SELECT id, order_id, product_id, product_name, quantity, tax_rate, unit_price, total_tax, total_discount, total_price, created_at, updated_at
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

		err = rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
			&item.TaxRate,
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
func (r *orderRepository) FindByIdempotencyKey(
	ctx context.Context,
	idempotencyKey uuid.UUID,
) (*entity.Order, error) {
	// Get order
	orderQuery := `
		SELECT id, idempotency_key, checkout_session_id, customer_id, status, payment_gateway, currency, courier, destination, origin, package, shipping_cost, subtotal, total_tax, total_discount, total_price, created_at, updated_at
		FROM orders
		WHERE idempotency_key = $1
	`

	row := r.db.QueryRow(ctx, orderQuery, idempotencyKey)

	var (
		order                                                 entity.Order
		courierData, destinationData, originData, packageData []byte
	)

	err := row.Scan(
		&order.ID,
		&order.IdempotencyKey,
		&order.CheckoutSessionID,
		&order.CustomerID,
		&order.Status,
		&order.PaymentGateway,
		&order.Currency,
		&courierData,
		&destinationData,
		&originData,
		&packageData,
		&order.ShippingCost,
		&order.Subtotal,
		&order.TotalTax,
		&order.TotalDiscount,
		&order.TotalPrice,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.OrderNotFoundErrorMessage)
		}

		return nil, fmt.Errorf("failed to scan order: %w", err)
	}

	// Unmarshal JSONB fields
	if err = sonic.Unmarshal(courierData, &order.Courier); err != nil {
		return nil, fmt.Errorf("failed to unmarshal courier: %w", err)
	}

	if err = sonic.Unmarshal(destinationData, &order.Destination); err != nil {
		return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
	}

	if err = sonic.Unmarshal(originData, &order.Origin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
	}

	if err = sonic.Unmarshal(packageData, &order.Package); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}

	// Get order items
	const itemsQuery = `
		SELECT id, order_id, product_id, product_name, quantity, tax_rate, unit_price, total_tax, total_discount, total_price, created_at, updated_at
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

		err = rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
			&item.TaxRate,
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
func (r *orderRepository) FindByCustomerID(
	ctx context.Context,
	customerID uuid.UUID,
	limit, offset int64,
) ([]*entity.Order, error) {
	query := `
		SELECT id, idempotency_key, checkout_session_id, customer_id, status, payment_gateway, currency, courier, destination, origin, package, shipping_cost, subtotal, total_tax, total_discount, total_price, created_at, updated_at
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
		var (
			order                                                 entity.Order
			courierData, destinationData, originData, packageData []byte
		)

		err = rows.Scan(
			&order.ID,
			&order.IdempotencyKey,
			&order.CheckoutSessionID,
			&order.CustomerID,
			&order.Status,
			&order.PaymentGateway,
			&order.Currency,
			&courierData,
			&destinationData,
			&originData,
			&packageData,
			&order.ShippingCost,
			&order.Subtotal,
			&order.TotalTax,
			&order.TotalDiscount,
			&order.TotalPrice,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Unmarshal JSONB fields
		if err = sonic.Unmarshal(courierData, &order.Courier); err != nil {
			return nil, fmt.Errorf("failed to unmarshal courier: %w", err)
		}

		if err = sonic.Unmarshal(destinationData, &order.Destination); err != nil {
			return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
		}

		if err = sonic.Unmarshal(originData, &order.Origin); err != nil {
			return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
		}

		if err = sonic.Unmarshal(packageData, &order.Package); err != nil {
			return nil, fmt.Errorf("failed to unmarshal package: %w", err)
		}

		orders = append(orders, &order)
	}

	// Load items for each order
	for _, order := range orders {
		items, rowErr := r.loadOrderItems(ctx, order.ID)
		if rowErr != nil {
			return nil, fmt.Errorf("failed to load items for order %s: %w", order.ID, rowErr)
		}

		order.Items = items
	}

	return orders, nil
}

// FindAll retrieves all orders with optional pagination.
func (r *orderRepository) FindAll(
	ctx context.Context,
	limit, offset int64,
) ([]*entity.Order, error) {
	query := `
		SELECT id, idempotency_key, checkout_session_id, customer_id, status, payment_gateway, currency, courier, destination, origin, package, shipping_cost, subtotal, total_tax, total_discount, total_price, created_at, updated_at
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
		var (
			order                                                 entity.Order
			courierData, destinationData, originData, packageData []byte
		)

		err = rows.Scan(
			&order.ID,
			&order.IdempotencyKey,
			&order.CheckoutSessionID,
			&order.CustomerID,
			&order.Status,
			&order.PaymentGateway,
			&order.Currency,
			&courierData,
			&destinationData,
			&originData,
			&packageData,
			&order.ShippingCost,
			&order.Subtotal,
			&order.TotalTax,
			&order.TotalDiscount,
			&order.TotalPrice,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Unmarshal JSONB fields
		if err = sonic.Unmarshal(courierData, &order.Courier); err != nil {
			return nil, fmt.Errorf("failed to unmarshal courier: %w", err)
		}

		if err = sonic.Unmarshal(destinationData, &order.Destination); err != nil {
			return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
		}

		if err = sonic.Unmarshal(originData, &order.Origin); err != nil {
			return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
		}

		if err = sonic.Unmarshal(packageData, &order.Package); err != nil {
			return nil, fmt.Errorf("failed to unmarshal package: %w", err)
		}

		orders = append(orders, &order)
	}

	// Load items for each order
	for _, order := range orders {
		items, rowErr := r.loadOrderItems(ctx, order.ID)
		if rowErr != nil {
			return nil, fmt.Errorf("failed to load items for order %s: %w", order.ID, rowErr)
		}

		order.Items = items
	}

	return orders, nil
}

// Update updates an existing order.
func (r *orderRepository) Update(
	ctx context.Context,
	order *entity.Order,
) (*entity.Order, error) {
	// Marshal JSONB fields
	courierJSON, err := sonic.Marshal(order.Courier)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal courier: %w", err)
	}

	destinationJSON, err := sonic.Marshal(order.Destination)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal destination: %w", err)
	}

	originJSON, err := sonic.Marshal(order.Origin)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal origin: %w", err)
	}

	packageJSON, err := sonic.Marshal(order.Package)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package: %w", err)
	}

	// Update the order itself
	updateOrderQuery := `
		UPDATE orders
		SET customer_id = $1,
			idempotency_key = $2,
			checkout_session_id = $3,
			status = $4,
			payment_gateway = $5,
			currency = $6,
			courier = $7,
			destination = $8,
			origin = $9,
			package = $10,
			shipping_cost = $11,
			subtotal = $12,
			total_tax = $13,
			total_discount = $14,
			total_price = $15,
			updated_at = $16
		WHERE id = $17
		RETURNING id, idempotency_key, checkout_session_id, customer_id, status, payment_gateway, currency, courier, destination, origin, package, shipping_cost, subtotal, total_tax, total_discount, total_price, created_at, updated_at
	`

	row := r.db.QueryRow(
		ctx,
		updateOrderQuery,
		order.CustomerID,        // $1
		order.IdempotencyKey,    // $2
		order.CheckoutSessionID, // $3
		order.Status,            // $4
		order.PaymentGateway,    // $5
		order.Currency,          // $6
		courierJSON,             // $7
		destinationJSON,         // $8
		originJSON,              // $9
		packageJSON,             // $10
		order.ShippingCost,      // $11
		order.Subtotal,          // $12
		order.TotalTax,          // $13
		order.TotalDiscount,     // $14
		order.TotalPrice,        // $15
		order.UpdatedAt,         // $16
		order.ID,                // $17
	)

	var (
		updatedOrder                                          entity.Order
		courierData, destinationData, originData, packageData []byte
	)

	err = row.Scan(
		&updatedOrder.ID,
		&updatedOrder.IdempotencyKey,
		&updatedOrder.CheckoutSessionID,
		&updatedOrder.CustomerID,
		&updatedOrder.Status,
		&updatedOrder.PaymentGateway,
		&updatedOrder.Currency,
		&courierData,
		&destinationData,
		&originData,
		&packageData,
		&updatedOrder.ShippingCost,
		&updatedOrder.Subtotal,
		&updatedOrder.TotalTax,
		&updatedOrder.TotalDiscount,
		&updatedOrder.TotalPrice,
		&updatedOrder.CreatedAt,
		&updatedOrder.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.OrderNotFoundErrorMessage)
		}

		return nil, fmt.Errorf("failed to scan updated order: %w", err)
	}

	// Unmarshal JSONB fields
	if err = sonic.Unmarshal(courierData, &updatedOrder.Courier); err != nil {
		return nil, fmt.Errorf("failed to unmarshal courier: %w", err)
	}

	if err = sonic.Unmarshal(destinationData, &updatedOrder.Destination); err != nil {
		return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
	}

	if err = sonic.Unmarshal(originData, &updatedOrder.Origin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
	}

	if err = sonic.Unmarshal(packageData, &updatedOrder.Package); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}

	// Delete existing items
	_, err = r.db.Exec(ctx, "DELETE FROM order_items WHERE order_id = $1", order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing order items: %w", err)
	}

	// Insert new items if provided
	if len(order.Items) > 0 {
		insertItemQuery := `
			INSERT INTO order_items (id, order_id, product_id, quantity, tax_rate, unit_price, total_tax, total_discount, total_price, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`

		for i := range len(order.Items) {
			item := &order.Items[i]

			_, err = r.db.Exec(
				ctx,
				insertItemQuery,
				item.ID,
				order.ID,
				item.ProductID,
				item.Quantity,
				item.TaxRate,
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
func (r *orderRepository) Delete(ctx context.Context, id uuid.UUID) error {
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
func (r *orderRepository) UpdateStatus(
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
func (r *orderRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM orders WHERE id = $1)"

	var exists bool

	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check order existence: %w", err)
	}

	return exists, nil
}

// Count returns the total number of orders.
func (r *orderRepository) Count(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM orders"

	var count int64

	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}

	return count, nil
}

// CountByCustomer returns the total number of orders for a specific customer.
func (r *orderRepository) CountByCustomer(
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

// FindByCustomerIDWithCursor retrieves orders for a customer with cursor-based pagination.
func (r *orderRepository) FindByCustomerIDWithCursor(
	ctx context.Context,
	customerID uuid.UUID,
	limit int64,
	cursorID string,
	cursorTimestamp int64,
) ([]*entity.Order, error) {
	var query string

	var args []interface{}

	// If cursor is provided, use it for pagination
	if cursorID != "" && cursorTimestamp > 0 {
		query = `
			SELECT id, idempotency_key, checkout_session_id, customer_id, status, payment_gateway, currency, courier, destination, origin, package, shipping_cost, subtotal, total_tax, total_discount, total_price, created_at, updated_at
			FROM orders
			WHERE customer_id = $1
			  AND (EXTRACT(EPOCH FROM created_at) < $2 OR (EXTRACT(EPOCH FROM created_at) = $2 AND id < $3))
			ORDER BY created_at DESC, id DESC
			LIMIT $4
		`
		args = []interface{}{customerID, cursorTimestamp, cursorID, limit}
	} else {
		query = `
			SELECT id, idempotency_key, checkout_session_id, customer_id, status, payment_gateway, currency, courier, destination, origin, package, shipping_cost, subtotal, total_tax, total_discount, total_price, created_at, updated_at
			FROM orders
			WHERE customer_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		`
		args = []interface{}{customerID, limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders with cursor: %w", err)
	}
	defer rows.Close()

	var orders []*entity.Order

	for rows.Next() {
		var (
			order                                                 entity.Order
			courierData, destinationData, originData, packageData []byte
		)

		err = rows.Scan(
			&order.ID,
			&order.IdempotencyKey,
			&order.CheckoutSessionID,
			&order.CustomerID,
			&order.Status,
			&order.PaymentGateway,
			&order.Currency,
			&courierData,
			&destinationData,
			&originData,
			&packageData,
			&order.ShippingCost,
			&order.Subtotal,
			&order.TotalTax,
			&order.TotalDiscount,
			&order.TotalPrice,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Unmarshal JSONB fields
		if err = sonic.Unmarshal(courierData, &order.Courier); err != nil {
			return nil, fmt.Errorf("failed to unmarshal courier: %w", err)
		}

		if err = sonic.Unmarshal(destinationData, &order.Destination); err != nil {
			return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
		}

		if err = sonic.Unmarshal(originData, &order.Origin); err != nil {
			return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
		}

		if err = sonic.Unmarshal(packageData, &order.Package); err != nil {
			return nil, fmt.Errorf("failed to unmarshal package: %w", err)
		}

		orders = append(orders, &order)
	}

	// Load items for each order
	for _, order := range orders {
		items, rowErr := r.loadOrderItems(ctx, order.ID)
		if rowErr != nil {
			return nil, fmt.Errorf("failed to load items for order %s: %w", order.ID, rowErr)
		}

		order.Items = items
	}

	return orders, nil
}

// FindAllWithCursor retrieves all orders with cursor-based pagination.
func (r *orderRepository) FindAllWithCursor(
	ctx context.Context,
	limit int64,
	cursorID string,
	cursorTimestamp int64,
) ([]*entity.Order, error) {
	var query string

	var args []interface{}

	// If cursor is provided, use it for pagination
	if cursorID != "" && cursorTimestamp > 0 {
		query = `
			SELECT id, idempotency_key, checkout_session_id, customer_id, status, payment_gateway, currency, courier, destination, origin, package, shipping_cost, subtotal, total_tax, total_discount, total_price, created_at, updated_at
			FROM orders
			WHERE (EXTRACT(EPOCH FROM created_at) < $1 OR (EXTRACT(EPOCH FROM created_at) = $1 AND id < $2))
			ORDER BY created_at DESC, id DESC
			LIMIT $3
		`
		args = []interface{}{cursorTimestamp, cursorID, limit}
	} else {
		query = `
			SELECT id, idempotency_key, checkout_session_id, customer_id, status, payment_gateway, currency, courier, destination, origin, package, shipping_cost, subtotal, total_tax, total_discount, total_price, created_at, updated_at
			FROM orders
			ORDER BY created_at DESC, id DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query all orders with cursor: %w", err)
	}
	defer rows.Close()

	var orders []*entity.Order

	for rows.Next() {
		var (
			order                                                 entity.Order
			courierData, destinationData, originData, packageData []byte
		)

		err = rows.Scan(
			&order.ID,
			&order.IdempotencyKey,
			&order.CheckoutSessionID,
			&order.CustomerID,
			&order.Status,
			&order.PaymentGateway,
			&order.Currency,
			&courierData,
			&destinationData,
			&originData,
			&packageData,
			&order.ShippingCost,
			&order.Subtotal,
			&order.TotalTax,
			&order.TotalDiscount,
			&order.TotalPrice,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Unmarshal JSONB fields
		if err = sonic.Unmarshal(courierData, &order.Courier); err != nil {
			return nil, fmt.Errorf("failed to unmarshal courier: %w", err)
		}

		if err = sonic.Unmarshal(destinationData, &order.Destination); err != nil {
			return nil, fmt.Errorf("failed to unmarshal destination: %w", err)
		}

		if err = sonic.Unmarshal(originData, &order.Origin); err != nil {
			return nil, fmt.Errorf("failed to unmarshal origin: %w", err)
		}

		if err = sonic.Unmarshal(packageData, &order.Package); err != nil {
			return nil, fmt.Errorf("failed to unmarshal package: %w", err)
		}

		orders = append(orders, &order)
	}

	// Load items for each order
	for _, order := range orders {
		items, rowErr := r.loadOrderItems(ctx, order.ID)
		if rowErr != nil {
			return nil, fmt.Errorf("failed to load items for order %s: %w", order.ID, rowErr)
		}

		order.Items = items
	}

	return orders, nil
}

// loadOrderItems is a helper method to load items for an order.
func (r *orderRepository) loadOrderItems(
	ctx context.Context,
	orderID uuid.UUID,
) ([]entity.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, product_name, quantity, tax_rate, unit_price, total_tax, total_discount, total_price, created_at, updated_at
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

		err = rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
			&item.TaxRate,
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
