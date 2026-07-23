-- name: GetItem :one
SELECT * FROM item
WHERE id = $1 LIMIT 1;

-- name: GetItemByName :one
SELECT * FROM item
WHERE name = $1 LIMIT 1;

-- name: SummarizeItem :one
SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id
FROM item
LEFT JOIN category ON (item.category_id = category.id)
LEFT JOIN item_list ON (item_list.item_id = item.id)
WHERE item.id = $1;

-- name: ListItems :many
SELECT * FROM item
ORDER BY name;

-- name: ListItemsForCategory :many
SELECT * FROM item
WHERE category_id = $1
ORDER BY name;

-- name: SummarizeItems :many
SELECT item.id, item.name, item.category_id, category.name AS category_name,
	item_list.id AS list_id, item_list.quantity AS list_quantity, item_list.done AS list_done
FROM item
LEFT JOIN category ON (item.category_id = category.id)
LEFT JOIN item_list ON (item_list.item_id = item.id)
ORDER BY category.name, item.name;

-- name: CreateItem :one
INSERT INTO item (category_id, name)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateItem :one
UPDATE item SET
  category_id = $2,
  name = $3
WHERE id = $1
RETURNING *;

-- name: DeleteItem :exec
DELETE FROM item
WHERE id = $1;
