-- name: GetListItem :one
SELECT item_list.id, item_list.item_id, item.name, item.category_id, item_list.quantity, item_list.done
FROM item_list
INNER JOIN item ON (item_list.item_id = item.id)
WHERE item_list.id = $1;

-- name: LoadList :many
SELECT item.id, item.name, item.category_id, category.name AS category_name,
	item_list.quantity AS list_quantity, item_list.id AS list_id, item_list.done AS list_done
FROM item_list
INNER JOIN item ON (item.id = item_list.item_id)
INNER JOIN category ON (item.category_id = category.id)
ORDER BY category.name, item.name;

-- name: CreateListItem :one
INSERT INTO item_list (item_id, quantity)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateListItem :one
UPDATE item_list SET
    quantity = $2,
    done = $3
WHERE item_id = $1
RETURNING *;

-- name: MarkItemDone :one
UPDATE item_list SET done = $2
WHERE item_id = $1
RETURNING *;

-- name: DeleteListItem :exec
DELETE FROM item_list
WHERE id = $1;

-- name: FinishShopping :exec
DELETE FROM item_list
WHERE done = TRUE;
