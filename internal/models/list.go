package models

import (
	"context"
	"errors"
)

type ListItem struct {
	ID         int
	ItemID     int
	CategoryID string
	Quantity   string
	Done       bool

	Name string
}

func (i *ListItem) Validate(ctx context.Context) error {
	return nil
}

func LoadList(ctx context.Context) ([]Item, error) {
	rows, err := db.QueryContext(ctx, `
SELECT item.id, item.name, item.category_id, category.name AS category_name,
	item_list.quantity AS list_quantity, item_list.id AS list_id, item_list.done AS list_done
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
		if err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.categoryName, &item.List.Quantity, &item.List.ID, &item.List.Done); err != nil {
			return nil, err
		}
		ret = append(ret, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func GetListItem(ctx context.Context, id int) (*ListItem, error) {
	if id == 0 {
		return nil, errors.New("not a valid item")
	}

	row := db.QueryRowContext(ctx, `
SELECT item_list.id, item_list.item_id, item.name, item.category_id, item_list.quantity, item_list.done
FROM item_list
INNER JOIN item ON (item_list.item_id = item.id)
WHERE item_list.id = $1`, id)
	if row.Err() != nil {
		return nil, row.Err()
	}

	ret := &ListItem{}
	if err := row.Scan(&ret.ID, &ret.ItemID, &ret.Name, &ret.CategoryID, &ret.Quantity, &ret.Done); err != nil {
		return nil, err
	}

	return ret, nil
}

func ListAddItem(ctx context.Context, id int, quantity string) error {
	if id == 0 {
		return errors.New("not a valid item")
	}

	_, err := db.ExecContext(ctx, `INSERT INTO item_list (item_id, quantity) VALUES ($1, $2)`, id, quantity)
	return err
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
