package models

import (
	"context"
)

type ListItem struct {
	Quantity string
	Done     bool
}

func LoadList(ctx context.Context) ([]Item, error) {
	rows, err := db.QueryContext(ctx, `
SELECT item.id, item.name, item.category_id, category.name AS category_name,
	item_list.quantity AS list_quantity, item_list.done AS list_done
FROM item_list
INNER JOIN item ON (item.id = item_list.item_id)
INNER JOIN category ON (item.category_id = category.id)
ORDER BY category.name, item.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Item{}
	for rows.Next() {
		item := Item{
			List: &ListItem{},
		}
		if err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.categoryName, &item.List.Quantity, &item.List.Done); err != nil {
			return nil, err
		}
		ret = append(ret, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func MarkItemDone(ctx context.Context, id string, value bool) error {
	_, err := db.ExecContext(ctx,
		`UPDATE item_list SET done = $2 WHERE item_id = $1`,
		id,
		value,
	)

	return err
}

func DeleteFromList(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM item_list WHERE item_id = $1`, id)
	return err
}

func FinishShopping(ctx context.Context) error {
	_, err := db.ExecContext(ctx, "DELETE FROM item_list WHERE done = TRUE")
	return err
}
