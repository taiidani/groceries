package models

import (
	"context"
	"database/sql"
)

const ListDBKey = "list"

type List struct {
	Categories []*Category
	db         *sql.DB
}

type Category struct {
	ID    string
	Name  string
	Items []Item
}

type Item struct {
	ID         string
	CategoryID string
	Name       string
	Quantity   string
	InBag      bool // The bag denotes in-progress item additions
	Done       bool
}

func NewList(db *sql.DB) *List {
	return &List{db: db}
}

func (l *List) LoadCategories(ctx context.Context) ([]Category, error) {
	rows, err := l.db.QueryContext(ctx, "SELECT id, name FROM category ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Category{}
	for rows.Next() {
		// Load the category
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, err
		}

		// Then its items
		cat.Items, err = l.LoadItemsForCategory(ctx, cat.ID)
		if err != nil {
			return nil, err
		}

		ret = append(ret, cat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (l *List) LoadItemsForCategory(ctx context.Context, id string) ([]Item, error) {
	rows, err := l.db.QueryContext(ctx,
		"SELECT id, name, quantity, in_bag, done FROM item WHERE category_id = $1 ORDER BY name",
		id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Item{}
	for rows.Next() {
		// Load the item
		var item Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &item.InBag, &item.Done); err != nil {
			return nil, err
		}
		ret = append(ret, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (l *List) AddCategory(ctx context.Context, name string) error {
	_, err := l.db.ExecContext(ctx, "INSERT INTO category (name) VALUES ($1)", name)
	return err
}

func (l *List) DeleteCategory(ctx context.Context, id string) error {
	_, err := l.db.ExecContext(ctx, "DELETE FROM item WHERE category_id = $1", id)
	if err != nil {
		return err
	}

	_, err = l.db.ExecContext(ctx, "DELETE FROM category WHERE id = $1", id)
	return err
}

func (l *List) AddItem(ctx context.Context, item Item) error {
	_, err := l.db.ExecContext(ctx,
		"INSERT INTO item (category_id, name, quantity, in_bag) VALUES ($1, $2, $3, $4)",
		item.CategoryID,
		item.Name,
		item.Quantity,
		item.InBag,
	)

	return err
}

func (l *List) DeleteItem(ctx context.Context, id string) error {
	_, err := l.db.ExecContext(ctx, "DELETE FROM item WHERE id = $1", id)
	return err
}

func (l *List) MarkItemDone(ctx context.Context, id string) error {
	_, err := l.db.ExecContext(ctx,
		"UPDATE item SET done = TRUE WHERE id = $1",
		id,
	)

	return err
}

func (l *List) SaveBag(ctx context.Context) error {
	_, err := l.db.ExecContext(ctx, "UPDATE item SET in_bag = FALSE")
	return err
}

func (l *List) FinishShopping(ctx context.Context) error {
	_, err := l.db.ExecContext(ctx, "DELETE FROM item WHERE done = TRUE")
	return err
}
