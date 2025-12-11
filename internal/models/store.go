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
	if len(c.Name) < 3 {
		return errors.New("provided name needs to be at least 3 characters")
	}

	// Check for existing Store
	if c.ID == 0 {
		existing, err := LoadStores(ctx)
		if err != nil {
			return fmt.Errorf("could not load stores: %w", err)
		}
		for _, data := range existing {
			if data.Name == c.Name {
				return errors.New("store already found")
			}
		}
	}

	return nil
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
	_, err := db.ExecContext(ctx, "INSERT INTO store (name) VALUES ($1)", data.Name)
	return err
}

func EditStore(ctx context.Context, data Store) error {
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
