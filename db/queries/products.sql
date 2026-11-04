
-- name: CreateProduct :one
INSERT INTO products (
    name,
    sku,
    description,
    price,
    stock_quantity,
    status
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetProductByID :one
SELECT *
FROM products
WHERE id = $1
LIMIT 1;

-- name: GetProductBySKU :one
SELECT *
FROM products
WHERE sku = $1
LIMIT 1;


-- name: ListProducts :many
SELECT *
FROM products
ORDER BY id DESC
LIMIT $1
OFFSET $2;


-- name: CountProducts :one
SELECT COUNT(*)
FROM products;


-- name: UpdateProduct :one
UPDATE products
SET
    name = $1,
    sku = $2,
    description = $3,
    price = $4,
    stock_quantity = $5,
    status = $6,
    updated_at = NOW()
WHERE id = $7
RETURNING *;


-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;

