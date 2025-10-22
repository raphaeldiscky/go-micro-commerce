// Package repository defines the interface for cart data operations.
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

// CartRepository defines the interface for cart data operations.
type CartRepository interface {
	// Create saves a new cart
	Create(ctx context.Context, cart *entity.Cart) (*entity.Cart, error)

	// FindActiveCartByUserID retrieves an active cart by its UserID
	FindActiveCartByUserID(ctx context.Context, userID uuid.UUID) (*entity.Cart, error)

	// FindByID retrieves a cart by its ID
	FindByID(ctx context.Context, cartID uuid.UUID) (*entity.Cart, error)

	// FindByUserID retrieves a cart by its UserID
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.Cart, error)

	// FindByUserIDForCheckout retrieves a cart by its UserID for checkout
	FindByUserIDForCheckout(ctx context.Context, userID uuid.UUID) (*entity.Cart, error)

	// FindCheckedOutCartWithUnselectedItems retrieves a checked out cart with ONLY unselected items
	FindCheckedOutCartWithUnselectedItems(
		ctx context.Context,
		userID uuid.UUID,
	) (*entity.Cart, error)

	// Update updates an existing cart
	Update(ctx context.Context, cart *entity.Cart) (*entity.Cart, error)

	// AddItem adds an item to the cart
	AddItem(ctx context.Context, cartID uuid.UUID, item *entity.CartItem) error

	// RemoveItem removes an item from the cart
	RemoveItem(ctx context.Context, cartID uuid.UUID, itemID uuid.UUID) error

	// UpdateItemQuantity updates the quantity of a cart item
	UpdateItemQuantity(
		ctx context.Context,
		cartID uuid.UUID,
		itemID uuid.UUID,
		quantity int64,
	) error

	// SelectForCheckout marks an item as selected for checkout
	SelectForCheckout(ctx context.Context, cartID uuid.UUID, itemID uuid.UUID, selected bool) error

	// UpdateStatus updates the status of a cart
	UpdateStatus(ctx context.Context, cartID uuid.UUID, status constant.CartStatus) error
}

const (
	// SQL query to find cart by customer_id.
	findCartByCustomerIDQuery = `
		SELECT id, customer_id, status, created_at, updated_at
		FROM carts
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	// SQL query to find active cart by customer_id.
	findActiveCartByCustomerIDQuery = `
		SELECT id, customer_id, status, created_at, updated_at
		FROM carts
		WHERE customer_id = $1 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1
	`

	// SQL query to find checked out cart by customer_id.
	findCheckedOutCartByCustomerIDQuery = `
		SELECT id, customer_id, status, created_at, updated_at
		FROM carts
		WHERE customer_id = $1 AND status = 'checked_out'
		ORDER BY created_at DESC
		LIMIT 1
	`

	// SQL query to update cart timestamp.
	updateCartTimestampQuery = `
		UPDATE carts
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
)

// cartRepository implements the CartRepository interface for PostgreSQL.
type cartRepository struct {
	db DBTX
}

// NewCartRepository creates a new instance of cartRepository.
func NewCartRepository(db DBTX) CartRepository {
	return &cartRepository{
		db: db,
	}
}

// Create creates a new cart in the database.
func (r *cartRepository) Create(
	ctx context.Context,
	cart *entity.Cart,
) (*entity.Cart, error) {
	// Insert cart
	insertCartQuery := `
        INSERT INTO carts (id, customer_id, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, customer_id, status, created_at, updated_at
    `

	var createdCart entity.Cart

	err := r.db.QueryRow(
		ctx,
		insertCartQuery,
		cart.ID,
		cart.CustomerID,
		cart.Status,
		cart.CreatedAt,
		cart.UpdatedAt,
	).Scan(
		&createdCart.ID,
		&createdCart.CustomerID,
		&createdCart.Status,
		&createdCart.CreatedAt,
		&createdCart.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cart: %w", err)
	}

	// Insert cart items if present
	if len(cart.Items) > 0 {
		const insertItemQuery = `
            INSERT INTO cart_items (id, cart_id, product_id, quantity, selected_for_checkout, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
        `

		for i := range len(cart.Items) {
			item := &cart.Items[i]

			_, err = r.db.Exec(
				ctx,
				insertItemQuery,
				item.ID,
				createdCart.ID,
				item.ProductID,
				item.Quantity,
				item.SelectedForCheckout,
				item.CreatedAt,
				item.UpdatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to insert cart item: %w", err)
			}
		}
	}

	createdCart.Items = cart.Items

	return &createdCart, nil
}

// FindByID retrieves a cart by its ID.
func (r *cartRepository) FindByID(
	ctx context.Context,
	cartID uuid.UUID,
) (*entity.Cart, error) {
	// Get cart
	cartQuery := `
		SELECT id, customer_id, status, created_at, updated_at
		FROM carts
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, cartQuery, cartID)

	var cart entity.Cart

	err := row.Scan(
		&cart.ID,
		&cart.CustomerID,
		&cart.Status,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.CartNotFoundErrorMessage)
		}

		return nil, fmt.Errorf("failed to scan cart: %w", err)
	}

	// Get cart items
	const itemsQuery = `
		SELECT id, cart_id, product_id, quantity, selected_for_checkout, created_at, updated_at
		FROM cart_items
		WHERE cart_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, itemsQuery, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart items: %w", err)
	}
	defer rows.Close()

	var items []entity.CartItem

	for rows.Next() {
		var item entity.CartItem

		err = rows.Scan(
			&item.ID,
			&item.CartID,
			&item.ProductID,
			&item.Quantity,
			&item.SelectedForCheckout,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}

		items = append(items, item)
	}

	cart.Items = items

	return &cart, nil
}

// FindActiveCartByUserID retrieves an active cart by its customer ID.
func (r *cartRepository) FindActiveCartByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*entity.Cart, error) {
	// Get active cart only
	row := r.db.QueryRow(ctx, findActiveCartByCustomerIDQuery, userID)

	var cart entity.Cart

	err := row.Scan(
		&cart.ID,
		&cart.CustomerID,
		&cart.Status,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.CartNotFoundErrorMessage)
		}

		return nil, fmt.Errorf("failed to scan cart: %w", err)
	}

	// Get all cart items
	const itemsQuery = `
		SELECT id, cart_id, product_id, quantity, selected_for_checkout, created_at, updated_at
		FROM cart_items
		WHERE cart_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, itemsQuery, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart items: %w", err)
	}
	defer rows.Close()

	var items []entity.CartItem

	for rows.Next() {
		var item entity.CartItem

		err = rows.Scan(
			&item.ID,
			&item.CartID,
			&item.ProductID,
			&item.Quantity,
			&item.SelectedForCheckout,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}

		items = append(items, item)
	}

	cart.Items = items

	return &cart, nil
}

// FindByUserID retrieves a cart by its customer ID.
func (r *cartRepository) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*entity.Cart, error) {
	// Get cart
	row := r.db.QueryRow(ctx, findCartByCustomerIDQuery, userID)

	var cart entity.Cart

	err := row.Scan(
		&cart.ID,
		&cart.CustomerID,
		&cart.Status,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.CartNotFoundErrorMessage)
		}

		return nil, fmt.Errorf("failed to scan cart: %w", err)
	}

	// Get cart items
	const itemsQuery = `
		SELECT id, cart_id, product_id, quantity, selected_for_checkout, created_at, updated_at
		FROM cart_items
		WHERE cart_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, itemsQuery, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart items: %w", err)
	}
	defer rows.Close()

	var items []entity.CartItem

	for rows.Next() {
		var item entity.CartItem

		err = rows.Scan(
			&item.ID,
			&item.CartID,
			&item.ProductID,
			&item.Quantity,
			&item.SelectedForCheckout,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}

		items = append(items, item)
	}

	cart.Items = items

	return &cart, nil
}

// FindByUserIDForCheckout retrieves a cart by its customer ID with ONLY selected items for checkout.
// This is optimized for checkout session creation - filters at database level.
func (r *cartRepository) FindByUserIDForCheckout(
	ctx context.Context,
	userID uuid.UUID,
) (*entity.Cart, error) {
	// Get active cart only
	row := r.db.QueryRow(ctx, findActiveCartByCustomerIDQuery, userID)

	var cart entity.Cart

	err := row.Scan(
		&cart.ID,
		&cart.CustomerID,
		&cart.Status,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.CartNotFoundErrorMessage)
		}

		return nil, fmt.Errorf("failed to scan cart: %w", err)
	}

	// Get ONLY selected cart items (database-level filtering)
	const itemsQuery = `
		SELECT id, cart_id, product_id, quantity, selected_for_checkout, created_at, updated_at
		FROM cart_items
		WHERE cart_id = $1 AND selected_for_checkout = true
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, itemsQuery, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query selected cart items: %w", err)
	}
	defer rows.Close()

	var items []entity.CartItem

	for rows.Next() {
		var item entity.CartItem

		err = rows.Scan(
			&item.ID,
			&item.CartID,
			&item.ProductID,
			&item.Quantity,
			&item.SelectedForCheckout,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}

		items = append(items, item)
	}

	cart.Items = items

	return &cart, nil
}

// FindCheckedOutCartWithUnselectedItems retrieves a checked out cart with ONLY unselected items.
// This is optimized for cart migration after order placement - filters at database level.
func (r *cartRepository) FindCheckedOutCartWithUnselectedItems(
	ctx context.Context,
	userID uuid.UUID,
) (*entity.Cart, error) {
	// Get checked out cart only
	row := r.db.QueryRow(ctx, findCheckedOutCartByCustomerIDQuery, userID)

	var cart entity.Cart

	err := row.Scan(
		&cart.ID,
		&cart.CustomerID,
		&cart.Status,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.CartNotFoundErrorMessage)
		}

		return nil, fmt.Errorf("failed to scan cart: %w", err)
	}

	// Get ONLY unselected cart items (database-level filtering)
	const itemsQuery = `
		SELECT id, cart_id, product_id, quantity, selected_for_checkout, created_at, updated_at
		FROM cart_items
		WHERE cart_id = $1 AND selected_for_checkout = false
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, itemsQuery, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unselected cart items: %w", err)
	}
	defer rows.Close()

	var items []entity.CartItem

	for rows.Next() {
		var item entity.CartItem

		err = rows.Scan(
			&item.ID,
			&item.CartID,
			&item.ProductID,
			&item.Quantity,
			&item.SelectedForCheckout,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}

		items = append(items, item)
	}

	cart.Items = items

	return &cart, nil
}

// Update updates an existing cart in the database.
func (r *cartRepository) Update(
	ctx context.Context,
	cart *entity.Cart,
) (*entity.Cart, error) {
	// Update cart
	updateCartQuery := `
		UPDATE carts
		SET status = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, customer_id, status, created_at, updated_at
	`

	var updatedCart entity.Cart

	err := r.db.QueryRow(
		ctx,
		updateCartQuery,
		cart.Status,
		cart.UpdatedAt,
		cart.ID,
	).Scan(
		&updatedCart.ID,
		&updatedCart.CustomerID,
		&updatedCart.Status,
		&updatedCart.CreatedAt,
		&updatedCart.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(constant.CartNotFoundErrorMessage)
		}

		return nil, fmt.Errorf("failed to update cart: %w", err)
	}

	updatedCart.Items = cart.Items

	return &updatedCart, nil
}

// AddItem adds an item to the cart.
func (r *cartRepository) AddItem(
	ctx context.Context,
	cartID uuid.UUID,
	item *entity.CartItem,
) error {
	// Check if item already exists (by product_id)
	checkQuery := `
		SELECT id, quantity
		FROM cart_items
		WHERE cart_id = $1 AND product_id = $2
	`

	var existingID uuid.UUID

	var existingQuantity int64

	err := r.db.QueryRow(ctx, checkQuery, cartID, item.ProductID).
		Scan(&existingID, &existingQuantity)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to check existing item: %w", err)
	}

	if err == nil {
		// Item exists, update quantity
		updateQuery := `
			UPDATE cart_items
			SET quantity = $1
			WHERE id = $2
		`

		_, err = r.db.Exec(ctx, updateQuery, existingQuantity+item.Quantity, existingID)
		if err != nil {
			return fmt.Errorf("failed to update item quantity: %w", err)
		}
	} else {
		// Item doesn't exist, insert new
		insertQuery := `
			INSERT INTO cart_items (id, cart_id, product_id, quantity, selected_for_checkout, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err = r.db.Exec(
			ctx,
			insertQuery,
			item.ID,
			cartID,
			item.ProductID,
			item.Quantity,
			item.SelectedForCheckout,
			item.CreatedAt,
			item.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert cart item: %w", err)
		}
	}

	// Update cart updated_at
	_, err = r.db.Exec(ctx, updateCartTimestampQuery, cartID)
	if err != nil {
		return fmt.Errorf("failed to update cart timestamp: %w", err)
	}

	return nil
}

// RemoveItem removes an item from the cart.
func (r *cartRepository) RemoveItem(
	ctx context.Context,
	cartID uuid.UUID,
	itemID uuid.UUID,
) error {
	deleteQuery := `
		DELETE FROM cart_items
		WHERE id = $1 AND cart_id = $2
	`

	result, err := r.db.Exec(ctx, deleteQuery, itemID, cartID)
	if err != nil {
		return fmt.Errorf("failed to remove cart item: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("cart item not found")
	}

	// Update cart updated_at
	_, err = r.db.Exec(ctx, updateCartTimestampQuery, cartID)
	if err != nil {
		return fmt.Errorf("failed to update cart timestamp: %w", err)
	}

	return nil
}

// UpdateItemQuantity updates the quantity of a cart item.
func (r *cartRepository) UpdateItemQuantity(
	ctx context.Context,
	cartID uuid.UUID,
	itemID uuid.UUID,
	quantity int64,
) error {
	updateQuery := `
		UPDATE cart_items
		SET quantity = $1
		WHERE id = $2 AND cart_id = $3
	`

	result, err := r.db.Exec(ctx, updateQuery, quantity, itemID, cartID)
	if err != nil {
		return fmt.Errorf("failed to update item quantity: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("cart item not found")
	}

	// Update cart updated_at
	_, err = r.db.Exec(ctx, updateCartTimestampQuery, cartID)
	if err != nil {
		return fmt.Errorf("failed to update cart timestamp: %w", err)
	}

	return nil
}

// SelectForCheckout marks an item as selected for checkout.
func (r *cartRepository) SelectForCheckout(
	ctx context.Context,
	cartID uuid.UUID,
	itemID uuid.UUID,
	selected bool,
) error {
	updateQuery := `
		UPDATE cart_items
		SET selected_for_checkout = $1
		WHERE id = $2 AND cart_id = $3
	`

	result, err := r.db.Exec(ctx, updateQuery, selected, itemID, cartID)
	if err != nil {
		return fmt.Errorf("failed to update item selection: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("cart item not found")
	}

	// Update cart updated_at
	_, err = r.db.Exec(ctx, updateCartTimestampQuery, cartID)
	if err != nil {
		return fmt.Errorf("failed to update cart timestamp: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of a cart.
func (r *cartRepository) UpdateStatus(
	ctx context.Context,
	cartID uuid.UUID,
	status constant.CartStatus,
) error {
	updateQuery := `
		UPDATE carts
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, updateQuery, status, cartID)
	if err != nil {
		return fmt.Errorf("failed to update cart status: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("cart not found")
	}

	return nil
}
