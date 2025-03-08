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
	rows, err := l.db.QueryContext(ctx, "SELECT oid, name FROM category ORDER BY name")
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
		"SELECT oid, name, quantity, in_bag, done FROM item WHERE category_id = ? ORDER BY name",
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

func (l *List) AddCategory(ctx context.Context, name string) (int64, error) {
	result, err := l.db.ExecContext(ctx, "INSERT INTO category (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (l *List) DeleteCategory(ctx context.Context, id string) error {
	_, err := l.db.ExecContext(ctx, "DELETE FROM item WHERE category_id = ?", id)
	if err != nil {
		return err
	}

	_, err = l.db.ExecContext(ctx, "DELETE FROM category WHERE oid = ?", id)
	return err
}

func (l *List) AddItem(ctx context.Context, item Item) (int64, error) {
	result, err := l.db.ExecContext(ctx,
		"INSERT INTO item (category_id, name, quantity, in_bag) VALUES (?, ?, ?, ?)",
		item.CategoryID,
		item.Name,
		item.Quantity,
		item.InBag,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (l *List) DeleteItem(ctx context.Context, id string) error {
	_, err := l.db.ExecContext(ctx, "DELETE FROM item WHERE oid = ?", id)
	return err
}

func (l *List) MarkItemDone(ctx context.Context, id string) error {
	_, err := l.db.ExecContext(ctx,
		"UPDATE item SET done = TRUE WHERE oid = ?",
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
