package models

import (
	"context"
)

type Item struct {
	ID           int
	CategoryID   string
	Name         string
	Bag          *BagItem
	List         *ListItem
	categoryName string
}

func (i *Item) CategoryName() string {
	return i.categoryName
}

func LoadItems(ctx context.Context) ([]Item, error) {
	rows, err := db.QueryContext(ctx, `
SELECT item.id, item.name, item.category_id, category.name AS category_name,
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
		if err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.categoryName, &inBag, &inList); err != nil {
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

func GetItem(ctx context.Context, id string) (Item, error) {
	ret := Item{}
	var inBag *bool
	var inList *bool
	err := db.QueryRowContext(ctx, `
SELECT id, category_id, name, category.name AS category_name,
	(SELECT TRUE FROM item_bag WHERE item_bag.item_id = item.id) AS in_bag,
	(SELECT TRUE FROM item_list WHERE item_list.item_id = item.id) AS in_list
FROM item
LEFT JOIN category ON (item.category_id = category.id)
WHERE id = $1`, id).
		Scan(&ret.ID, &ret.CategoryID, &ret.Name, &ret.categoryName, &inBag, &inList)

	if inBag != nil {
		ret.Bag = &BagItem{}
	}
	if inList != nil {
		ret.List = &ListItem{}
	}

	return ret, err
}

func ItemChangeCategory(ctx context.Context, id int, categoryID int) error {
	_, err := db.ExecContext(ctx, `UPDATE item SET category_id = $2 WHERE id = $1`, id, categoryID)
	return err
}

func DeleteItem(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM item WHERE id = $1`, id)
	return err
}
