package models

import (
	"context"
	"errors"
)

type Item struct {
	ID           int
	CategoryID   string
	Name         string
	List         *ListItem
	categoryName string
}

func (i *Item) CategoryName() string {
	return i.categoryName
}

func (i *Item) Validate(ctx context.Context) error {
	return nil
}

func LoadItems(ctx context.Context) ([]Item, error) {
	rows, err := db.QueryContext(ctx, `
SELECT item.id, item.name, item.category_id, category.name AS category_name,
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
		var inList *bool
		if err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.categoryName, &inList); err != nil {
			return nil, err
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
	var inList *bool
	err := db.QueryRowContext(ctx, `
SELECT id, category_id, name, category.name AS category_name,
	(SELECT TRUE FROM item_list WHERE item_list.item_id = item.id) AS in_list
FROM item
LEFT JOIN category ON (item.category_id = category.id)
WHERE id = $1`, id).
		Scan(&ret.ID, &ret.CategoryID, &ret.Name, &ret.categoryName, &inList)

	if inList != nil {
		ret.List = &ListItem{}
	}

	return ret, err
}

func ItemChangeCategory(ctx context.Context, id int, categoryID int) error {
	_, err := db.ExecContext(ctx, `UPDATE item SET category_id = $2 WHERE id = $1`, id, categoryID)
	return err
}

func AddItem(ctx context.Context, item Item) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if item.ID == 0 {
		_, err = insertWithID(ctx, tx,
			`INSERT INTO item (category_id, name) VALUES ($1, $2) RETURNING id`,
			item.CategoryID,
			item.Name,
		)
		if err != nil {
			return errors.Join(tx.Rollback(), err)
		}
	}

	return tx.Commit()
}

func DeleteItem(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM item WHERE id = $1`, id)
	return err
}
