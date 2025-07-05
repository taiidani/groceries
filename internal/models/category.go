package models

import (
	"context"
	"errors"
	"fmt"
)

type Category struct {
	ID          string
	Name        string
	Description string
	ItemCount   int
}

const UncategorizedCategoryID string = "0"

func (c *Category) Validate(ctx context.Context) error {
	if len(c.Name) < 3 {
		return errors.New("provided name needs to be at least 3 characters")
	}

	// Check for existing category
	if c.ID == "" {
		existingCategories, err := LoadCategories(ctx)
		if err != nil {
			return fmt.Errorf("could not load categories: %w", err)
		}
		for _, cat := range existingCategories {
			if cat.Name == c.Name {
				return errors.New("category already found")
			}
		}
	}

	return nil
}

func LoadCategories(ctx context.Context) ([]Category, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, name, description, (SELECT COUNT(id) FROM item WHERE item.category_id = category.id)
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
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.ItemCount); err != nil {
			return nil, err
		}

		ret = append(ret, cat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func GetCategory(ctx context.Context, id string) (Category, error) {
	row := db.QueryRowContext(ctx, `
SELECT id, name, description,
 (SELECT COUNT(id) FROM item WHERE item.category_id = category.id)
FROM category
WHERE id = $1
ORDER BY name`, id)
	if row.Err() != nil {
		return Category{}, row.Err()
	}

	// Load the category
	var cat Category
	err := row.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.ItemCount)
	return cat, err
}

func AddCategory(ctx context.Context, cat Category) error {
	_, err := db.ExecContext(ctx, "INSERT INTO category (name, description) VALUES ($1, $2)", cat.Name, cat.Description)
	return err
}

func EditCategory(ctx context.Context, cat Category) error {
	_, err := db.ExecContext(ctx, `
UPDATE category SET
	name = $2,
	description = $3
WHERE id = $1`, cat.ID, cat.Name, cat.Description)
	return err
}

func DeleteCategory(ctx context.Context, id string) error {
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
