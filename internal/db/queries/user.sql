-- name: GetUser :one
SELECT * FROM "user"
WHERE id = $1 LIMIT 1;

-- name: GetUserByName :one
SELECT * FROM "user"
WHERE name = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM "user"
ORDER BY name;

-- name: CreateUser :one
INSERT INTO "user" (name, admin)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateUser :one
UPDATE "user" SET
  name = $2,
  admin = $3
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "user"
WHERE id = $1;

-- name: UsersForGroup :many
SELECT "user".* FROM user_group
JOIN "user" ON user_group.user_id = "user".id
WHERE group_id = $1;

-- name: AddUserToGroup :exec
INSERT INTO user_group (user_id, group_id)
VALUES ($1, $2);

-- name: RemoveUserFromGroup :exec
DELETE FROM user_group
WHERE user_id = $1 AND group_id = $2;
