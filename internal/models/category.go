package models

import (
	"context"
	"errors"
	"fmt"
)

type Category struct {
	ID          int
	StoreID     int
	Name        string
	Description string
	ItemCount   int
}

const UncategorizedCategoryID int = 0

func (c *Category) Store(ctx context.Context) (Store, error) {
	return GetStore(ctx, c.StoreID)
}

func (c *Category) Items(ctx context.Context) ([]Item, error) {
	items, err := LoadItems(ctx)
	if err != nil {
		return nil, err
	}

	var ret []Item
	for _, item := range items {
		if item.CategoryID == c.ID {
			ret = append(ret, item)
		}
	}

	return ret, nil
}

func (c *Category) Validate(ctx context.Context) error {
	var vErr error

	if len(c.Name) < 3 {
		vErr = errors.Join(vErr, errors.New("provided name needs to be at least 3 characters"))
	}

	if _, err := GetStore(ctx, c.StoreID); err != nil {
		vErr = errors.Join(vErr, fmt.Errorf("store not found: %w", err))
	}

	// Check for existing category
	if c.ID == 0 {
		if existingCategories, err := LoadCategories(ctx); err != nil {
			vErr = errors.Join(vErr, fmt.Errorf("could not load categories: %w", err))
		} else {
			for _, cat := range existingCategories {
				if cat.Name == c.Name && cat.StoreID == c.StoreID {
					vErr = errors.Join(vErr, errors.New("category already exists"))
				}
			}
		}
	}

	return vErr
}

func LoadCategories(ctx context.Context) ([]Category, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, store_id, name, description, (SELECT COUNT(id) FROM item WHERE item.category_id = category.id)
FROM category
ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Category{}
	for rows.Next() {
		// Load the category
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.StoreID, &cat.Name, &cat.Description, &cat.ItemCount); err != nil {
			return nil, err
		}

		ret = append(ret, cat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func GetCategory(ctx context.Context, id int) (Category, error) {
	row := db.QueryRowContext(ctx, `
SELECT id, store_id, name, description,
 (SELECT COUNT(id) FROM item WHERE item.category_id = category.id)
FROM category
WHERE id = $1
ORDER BY name`, id)
	if row.Err() != nil {
		return Category{}, row.Err()
	}

	// Load the category
	var cat Category
	err := row.Scan(&cat.ID, &cat.StoreID, &cat.Name, &cat.Description, &cat.ItemCount)
	if err != nil {
		return cat, err
	}

	return cat, err
}

func AddCategory(ctx context.Context, cat Category) error {
	if err := cat.Validate(ctx); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}

	_, err := db.ExecContext(ctx, "INSERT INTO category (name, store_id, description) VALUES ($1, $2, $3)", cat.Name, cat.StoreID, cat.Description)
	return err
}

func EditCategory(ctx context.Context, cat Category) error {
	if err := cat.Validate(ctx); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}

	_, err := db.ExecContext(ctx, `
UPDATE category SET
	name = $2,
	store_id = $3,
	description = $4
WHERE id = $1`, cat.ID, cat.Name, cat.StoreID, cat.Description)
	return err
}

func DeleteCategory(ctx context.Context, id int) error {
	// Prevent deletion if category is still in use
	items, err := LoadItems(ctx)
	if err != nil {
		return fmt.Errorf("could not enumerate item categories: %w", err)
	}

	for _, item := range items {
		if item.CategoryID == id {
			return errors.New("category is still in use")
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "DELETE FROM item WHERE category_id = $1", id)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	_, err = db.ExecContext(ctx, "DELETE FROM category WHERE id = $1", id)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	return tx.Commit()
}
