package models

import (
	"context"
	"errors"
	"fmt"
)

type BagItem struct {
	Quantity string
}

func LoadBag(ctx context.Context) ([]Item, error) {
	rows, err := db.QueryContext(ctx, `
SELECT item.id, item.name, item.category_id, category.name as category_name, item_bag.quantity AS bag_quantity
FROM item_bag
INNER JOIN item ON (item.id = item_bag.item_id)
INNER JOIN category ON (item.category_id = category.id)
ORDER BY category.name, item.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Item{}
	for rows.Next() {
		item := Item{
			Bag: &BagItem{},
		}
		if err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.categoryName, &item.Bag.Quantity); err != nil {
			return nil, err
		}
		ret = append(ret, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func AddItem(ctx context.Context, item Item) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	id := item.ID
	if item.ID == 0 {
		id, err = insertWithID(ctx, tx,
			`INSERT INTO item (category_id, name) VALUES ($1, $2) RETURNING id`,
			item.CategoryID,
			item.Name,
		)
		if err != nil {
			return errors.Join(tx.Rollback(), err)
		}
	}

	if item.Bag != nil {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO item_bag (item_id, quantity) VALUES ($1, $2)`,
			id,
			item.Bag.Quantity,
		); err != nil {
			return errors.Join(tx.Rollback(), err)
		}
	}

	return tx.Commit()
}

func AddExistingItem(ctx context.Context, id int, quantity string) error {
	if id == 0 {
		return errors.New("not a valid item")
	}

	_, err := db.ExecContext(ctx, `INSERT INTO item_bag (item_id, quantity) VALUES ($1, $2)`, id, quantity)
	return err
}

func BagUpdateItemName(ctx context.Context, id int, name, quantity string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE item SET name = $2 WHERE id = $1`,
		id,
		name,
	)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE item_bag SET quantity = $2 WHERE item_id = $1`,
		id,
		quantity,
	)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	return tx.Commit()
}

func UpdateItem(ctx context.Context, item Item) error {
	if item.ID == 0 {
		return errors.New("cannot update item with empty id")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE item SET category_id = $2, name = $3 WHERE id = $1`,
		item.ID,
		item.CategoryID,
		item.Name,
	)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE item_bag SET quantity = $2 WHERE item_id = $1`,
		item.ID,
		item.Bag.Quantity,
	)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	return tx.Commit()
}

func DeleteFromBag(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM item_bag WHERE item_id = $1`, id)
	return err
}

func SaveBag(ctx context.Context) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO item_list (item_id, quantity) SELECT item_id, quantity FROM item_bag`); err != nil {
		return errors.Join(tx.Rollback(), fmt.Errorf("unable to add bag items to list: %w", err))
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM item_bag`); err != nil {
		return errors.Join(tx.Rollback(), fmt.Errorf("unable to clear bag: %w", err))
	}

	return tx.Commit()
}
