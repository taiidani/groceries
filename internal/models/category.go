package models

import "context"

type Category struct {
	ID          string
	Name        string
	Description string
	ItemCount   int
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

func AddCategory(ctx context.Context, cat Category) error {
	_, err := db.ExecContext(ctx, "INSERT INTO category (name, description) VALUES ($1, $2)", cat.Name, cat.Description)
	return err
}

func DeleteCategory(ctx context.Context, id string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM item WHERE category_id = $1", id)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "DELETE FROM category WHERE id = $1", id)
	return err
}
