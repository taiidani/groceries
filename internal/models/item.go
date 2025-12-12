package models

import (
	"context"
	"errors"
	"fmt"
)

type Item struct {
	ID           int
	CategoryID   int
	Name         string
	List         *ListItem
	categoryName string
}

func (i *Item) CategoryName() string {
	return i.categoryName
}

func (i *Item) Category(ctx context.Context) (Category, error) {
	return GetCategory(ctx, i.CategoryID)
}

func (i *Item) Validate(ctx context.Context) error {
	var vErr error

	if _, err := GetCategory(ctx, i.CategoryID); err != nil {
		vErr = errors.Join(vErr, fmt.Errorf("category not found: %w", err))
	}

	if i.List != nil {
		vErr = errors.Join(vErr, i.List.Validate(ctx))
	}

	return vErr
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

func GetItem(ctx context.Context, id int) (Item, error) {
	ret := Item{}
	var listID *int
	err := db.QueryRowContext(ctx, `
SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id
FROM item
LEFT JOIN category ON (item.category_id = category.id)
LEFT JOIN item_list ON (item_list.item_id = item.id)
WHERE item.id = $1`, id).
		Scan(&ret.ID, &ret.CategoryID, &ret.Name, &ret.categoryName, &listID)
	if err != nil {
		return ret, err
	}

	if listID != nil {
		ret.List, err = GetListItem(ctx, *listID)
	}

	return ret, err
}

func GetItemByName(ctx context.Context, name string) (Item, error) {
	ret := Item{}
	var inList *bool
	err := db.QueryRowContext(ctx, `
SELECT item.id, item.category_id, item.name, category.name AS category_name,
	(SELECT TRUE FROM item_list WHERE item_list.item_id = item.id) AS in_list
FROM item
LEFT JOIN category ON (item.category_id = category.id)
WHERE item.name = $1`, name).
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

func AddItem(ctx context.Context, i Item) error {
	if err := i.Validate(ctx); err != nil {
		return fmt.Errorf("invalid item: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if i.ID == 0 {
		_, err = insertWithID(ctx, tx,
			`INSERT INTO item (category_id, name) VALUES ($1, $2) RETURNING id`,
			i.CategoryID,
			i.Name,
		)
		if err != nil {
			return errors.Join(tx.Rollback(), err)
		}
	}

	return tx.Commit()
}

func EditItem(ctx context.Context, i Item) error {
	if err := i.Validate(ctx); err != nil {
		return fmt.Errorf("invalid item: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
UPDATE item SET
	category_id = $2,
	name = $3
WHERE id = $1`, i.ID, i.CategoryID, i.Name)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	if i.List != nil {
		_, err := db.ExecContext(ctx, `
UPDATE item_list SET
	quantity = $2
WHERE id = $1`, i.List.ID, i.List.Quantity)
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
