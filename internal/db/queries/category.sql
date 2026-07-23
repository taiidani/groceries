-- name: GetCategory :one
SELECT *
FROM category
WHERE id = $1 LIMIT 1;

-- name: GetCategoryByName :one
SELECT *
FROM category
WHERE name = $1 LIMIT 1;

-- name: ListCategories :many
SELECT *
FROM category
ORDER BY name;

-- name: ListCategoriesForStore :many
SELECT *
FROM category
WHERE store_id = $1
ORDER BY name;

-- name: ListCategoriesWithItemCount :many
SELECT *, (SELECT COUNT(item.id) FROM item WHERE item.category_id = category.id) as item_count
FROM category
ORDER BY category.name;

-- name: CreateCategory :one
INSERT INTO category (name, store_id, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateCategory :one
UPDATE category SET
  name = $2,
  store_id = $3,
  description = $4
WHERE id = $1
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM category
WHERE id = $1;
