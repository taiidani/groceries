-- name: GetStore :one
SELECT * FROM store
WHERE id = $1 LIMIT 1;

-- name: GetStoreByName :one
SELECT * FROM store
WHERE name = $1 LIMIT 1;

-- name: ListStores :many
SELECT * FROM store
ORDER BY name;

-- name: CreateStore :one
INSERT INTO store (name)
VALUES ($1)
RETURNING *;

-- name: UpdateStore :one
UPDATE store SET
  name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteStore :exec
DELETE FROM store
WHERE id = $1;
