-- name: CreateProduct :one
INSERT INTO products (id, name, description, price, seller_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProducts :many
SELECT * FROM products
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetProductsBySeller :many
SELECT * FROM products
WHERE seller_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateProduct :one
UPDATE products
SET name = $2, description = $3, price = $4, seller_id = $5, updated_at = $6
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteProduct :exec
UPDATE products
SET deleted_at = NOW()
WHERE id = $1;
