-- name: CreateSeller :one
INSERT INTO sellers (id, name, email, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSellerByID :one
SELECT * FROM sellers
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListSellers :many
SELECT * FROM sellers
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateSeller :one
UPDATE sellers
SET name = $2, email = $3, updated_at = $4
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteSeller :exec
UPDATE sellers
SET deleted_at = NOW()
WHERE id = $1;
