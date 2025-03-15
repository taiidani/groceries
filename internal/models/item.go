package models

import (
	"context"
)

type Item struct {
	ID         int
	CategoryID string
	Name       string
	Bag        *BagItem
	List       *ListItem
}

func LoadItems(ctx context.Context) ([]Item, error) {
	rows, err := db.QueryContext(ctx, `
SELECT item.id, item.name, item.category_id,
	(SELECT TRUE FROM item_bag WHERE item_bag.item_id = item.id) AS in_bag,
	(SELECT TRUE FROM item_list WHERE item_list.item_id = item.id) AS in_list
FROM item
LEFT JOIN category ON (item.category_id = category.id)
ORDER BY category.name, item.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Item{}
	for rows.Next() {
		// Load the item
		item := Item{}
		var inBag *bool
		var inList *bool
		if err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &inBag, &inList); err != nil {
			return nil, err
		}

		if inBag != nil {
			item.Bag = &BagItem{}
		}
		if inList != nil {
			item.List = &ListItem{}
		}

		ret = append(ret, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func DeleteItem(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM item WHERE id = $1`, id)
	return err
}
