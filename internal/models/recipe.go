package models

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Recipe struct {
	ID          int
	Name        string
	Description string
	Items       []RecipeItem
	CreatedAt   time.Time
}

type RecipeItem struct {
	ID       int
	RecipeID int
	ItemID   int
	Quantity string
	ItemName string
	InList   bool
}

func (r *Recipe) Validate(ctx context.Context) error {
	var vErr error

	if r.Name == "" {
		vErr = errors.Join(vErr, errors.New("recipe name cannot be empty"))
	}

	return vErr
}

func LoadRecipes(ctx context.Context) ([]Recipe, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, name, description, created_at
FROM recipe
ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Recipe{}
	for rows.Next() {
		recipe := Recipe{}
		if err := rows.Scan(&recipe.ID, &recipe.Name, &recipe.Description, &recipe.CreatedAt); err != nil {
			return nil, err
		}
		ret = append(ret, recipe)
	}

	return ret, rows.Err()
}

func GetRecipe(ctx context.Context, id int) (Recipe, error) {
	ret := Recipe{}

	err := db.QueryRowContext(ctx, `
SELECT id, name, description, created_at
FROM recipe
WHERE id = $1`, id).
		Scan(&ret.ID, &ret.Name, &ret.Description, &ret.CreatedAt)
	if err != nil {
		return ret, err
	}

	// Load recipe items
	rows, err := db.QueryContext(ctx, `
SELECT recipe_item.id, recipe_item.recipe_id, recipe_item.item_id, recipe_item.quantity, item.name,
	(SELECT TRUE FROM item_list WHERE item_list.item_id = recipe_item.item_id) AS in_list
FROM recipe_item
INNER JOIN item ON (recipe_item.item_id = item.id)
WHERE recipe_item.recipe_id = $1
ORDER BY item.name`, id)
	if err != nil {
		return ret, err
	}
	defer rows.Close()

	ret.Items = []RecipeItem{}
	for rows.Next() {
		recipeItem := RecipeItem{}
		var inList *bool
		if err := rows.Scan(&recipeItem.ID, &recipeItem.RecipeID, &recipeItem.ItemID, &recipeItem.Quantity, &recipeItem.ItemName, &inList); err != nil {
			return ret, err
		}
		if inList != nil && *inList {
			recipeItem.InList = true
		}
		ret.Items = append(ret.Items, recipeItem)
	}
	if err := rows.Err(); err != nil {
		return ret, err
	}

	return ret, nil
}

func AddRecipe(ctx context.Context, r Recipe) (int, error) {
	if err := r.Validate(ctx); err != nil {
		return 0, fmt.Errorf("invalid recipe: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	id, err := insertWithID(ctx, tx,
		`INSERT INTO recipe (name, description) VALUES ($1, $2) RETURNING id`,
		r.Name,
		r.Description,
	)
	if err != nil {
		return 0, errors.Join(tx.Rollback(), err)
	}

	return id, tx.Commit()
}

func EditRecipe(ctx context.Context, r Recipe) error {
	if err := r.Validate(ctx); err != nil {
		return fmt.Errorf("invalid recipe: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
UPDATE recipe SET
	name = $2,
	description = $3
WHERE id = $1`, r.ID, r.Name, r.Description)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	return tx.Commit()
}

func DeleteRecipe(ctx context.Context, id int) error {
	_, err := db.ExecContext(ctx, `DELETE FROM recipe WHERE id = $1`, id)
	return err
}

func AddRecipeItem(ctx context.Context, recipeID int, itemID int, quantity string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO recipe_item (recipe_id, item_id, quantity) VALUES ($1, $2, $3)`,
		recipeID, itemID, quantity)
	return err
}

func RemoveRecipeItem(ctx context.Context, recipeID int, itemID int) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM recipe_item WHERE recipe_id = $1 AND item_id = $2`,
		recipeID, itemID)
	return err
}

func AddRecipeToList(ctx context.Context, recipeID int) error {
	// Get all items in the recipe
	recipe, err := GetRecipe(ctx, recipeID)
	if err != nil {
		return err
	}

	// Add each item to the list if it's not already there
	for _, recipeItem := range recipe.Items {
		// Check if item is already in list
		var exists bool
		err := db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM item_list WHERE item_id = $1)`,
			recipeItem.ItemID).Scan(&exists)
		if err != nil {
			return err
		}

		// Only add if not already in list
		if !exists {
			err = ListAddItem(ctx, recipeItem.ItemID, recipeItem.Quantity)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func GetRecipesUsingItem(ctx context.Context, itemID int) ([]Recipe, error) {
	rows, err := db.QueryContext(ctx, `
SELECT DISTINCT recipe.id, recipe.name, recipe.description, recipe.created_at
FROM recipe
INNER JOIN recipe_item ON (recipe.id = recipe_item.recipe_id)
WHERE recipe_item.item_id = $1
ORDER BY recipe.name`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Recipe{}
	for rows.Next() {
		recipe := Recipe{}
		if err := rows.Scan(&recipe.ID, &recipe.Name, &recipe.Description, &recipe.CreatedAt); err != nil {
			return nil, err
		}
		ret = append(ret, recipe)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}
