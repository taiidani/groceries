package models

import (
	"context"
	"errors"
	"fmt"
)

type Store struct {
	ID   int
	Name string
}

const UncategorizedStoreID int = 0

func (c *Store) Validate(ctx context.Context) error {
	var vErr error

	if len(c.Name) < 3 {
		vErr = errors.Join(vErr, errors.New("provided name needs to be at least 3 characters"))
	}

	// Check for existing Store
	if c.ID == 0 {
		existing, err := LoadStores(ctx)
		if err != nil {
			vErr = errors.Join(vErr, fmt.Errorf("could not load stores: %w", err))
		}
		for _, data := range existing {
			if data.Name == c.Name {
				vErr = errors.Join(vErr, errors.New("store already found"))
			}
		}
	}

	return vErr
}

func LoadStores(ctx context.Context) ([]Store, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, name
FROM store
ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Store{}
	for rows.Next() {
		// Load the store
		var data Store
		if err := rows.Scan(&data.ID, &data.Name); err != nil {
			return nil, err
		}

		ret = append(ret, data)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func GetStore(ctx context.Context, id int) (Store, error) {
	row := db.QueryRowContext(ctx, `
SELECT id, name
FROM store
WHERE id = $1`, id)
	if row.Err() != nil {
		return Store{}, row.Err()
	}

	// Load the store
	var data Store
	err := row.Scan(&data.ID, &data.Name)
	return data, err
}

func AddStore(ctx context.Context, data Store) error {
	if err := data.Validate(ctx); err != nil {
		return fmt.Errorf("invalid store: %w", err)
	}

	_, err := db.ExecContext(ctx, "INSERT INTO store (name) VALUES ($1)", data.Name)
	return err
}

func EditStore(ctx context.Context, data Store) error {
	if err := data.Validate(ctx); err != nil {
		return fmt.Errorf("invalid store: %w", err)
	}

	_, err := db.ExecContext(ctx, `
UPDATE store SET
	name = $2
WHERE id = $1`, data.ID, data.Name)
	return err
}

func DeleteStore(ctx context.Context, id int) error {
	// Prevent deletion if store is still in use
	cats, err := LoadCategories(ctx)
	if err != nil {
		return fmt.Errorf("could not enumerate store categories: %w", err)
	}

	for _, cat := range cats {
		if cat.StoreID == id {
			return errors.New("store is still in use")
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "DELETE FROM store WHERE id = $1", id)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	return tx.Commit()
}

func (s *Store) Categories(ctx context.Context) ([]Category, error) {
	categories, err := LoadCategories(ctx)
	if err != nil {
		return nil, err
	}

	var ret []Category
	for _, category := range categories {
		if category.StoreID == s.ID {
			ret = append(ret, category)
		}
	}

	return ret, nil
}
