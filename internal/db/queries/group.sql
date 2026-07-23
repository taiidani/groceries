-- name: GetGroup :one
SELECT * FROM "group"
WHERE id = $1 LIMIT 1;

-- name: GetGroupByName :one
SELECT * FROM "group"
WHERE name = $1 LIMIT 1;

-- name: ListGroups :many
SELECT * FROM "group"
ORDER BY name;

-- name: CreateGroup :one
INSERT INTO "group" (name)
VALUES ($1)
RETURNING *;

-- name: UpdateGroup :one
UPDATE "group" SET
  name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteGroup :exec
DELETE FROM "group"
WHERE id = $1;
