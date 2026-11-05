
-- name: CreateProduct :one
INSERT INTO products (
    name,
    sku,
    description,
    price,
    stock_quantity,
    status
)
VALUES ($1, $2, $3, $4, $5, $6)
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
WHERE
    (sqlc.arg(status)::varchar = '' OR status = sqlc.arg(status)::varchar)
    AND (
        sqlc.arg(search)::text = ''
        OR name ILIKE '%' || sqlc.arg(search)::text || '%'
    )
    AND (
        NOT sqlc.arg(low_stock)::boolean
        OR stock_quantity <= 5
    )
ORDER BY id DESC
LIMIT sqlc.arg(page_size)::integer
OFFSET sqlc.arg(page_offset)::integer;

-- name: CountProducts :one
SELECT COUNT(*)
FROM products
WHERE
    (sqlc.arg(status)::varchar = '' OR status = sqlc.arg(status)::varchar)
    AND (
        sqlc.arg(search)::text = ''
        OR name ILIKE '%' || sqlc.arg(search)::text || '%'
    )
    AND (
        NOT sqlc.arg(low_stock)::boolean
        OR stock_quantity <= 5
    );

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
