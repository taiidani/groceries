package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Item struct {
	ID           int       `json:"id"`
	CategoryID   int       `json:"category_id"`
	Name         string    `json:"name"`
	List         *ListItem `json:"list"`
	categoryName string
}

func (i Item) MarshalJSON() ([]byte, error) {
	type itemJSON struct {
		ID           int       `json:"id"`
		CategoryID   int       `json:"category_id"`
		CategoryName string    `json:"category_name"`
		Name         string    `json:"name"`
		List         *ListItem `json:"list"`
	}

	return json.Marshal(itemJSON{
		ID:           i.ID,
		CategoryID:   i.CategoryID,
		CategoryName: i.categoryName,
		Name:         i.Name,
		List:         i.List,
	})
}

func (i *Item) CategoryName() string {
	return i.categoryName
}

func (i *Item) Category(ctx context.Context) (Category, error) {
	return GetCategory(ctx, i.CategoryID)
}

func (i *Item) Validate(ctx context.Context) error {
	var vErr error

	if _, err := GetCategory(ctx, i.CategoryID); err != nil {
		vErr = errors.Join(vErr, fmt.Errorf("category not found: %w", err))
	}

	if i.List != nil {
		vErr = errors.Join(vErr, i.List.Validate(ctx))
	}

	return vErr
}

func LoadItems(ctx context.Context) ([]Item, error) {
	rows, err := db.QueryContext(ctx, `
SELECT item.id, item.name, item.category_id, category.name AS category_name,
	item_list.id AS list_id, item_list.quantity AS list_quantity, item_list.done AS list_done
FROM item
LEFT JOIN category ON (item.category_id = category.id)
LEFT JOIN item_list ON (item_list.item_id = item.id)
ORDER BY category.name, item.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []Item{}
	for rows.Next() {
		// Load the item
		item := Item{}
		var listID *int
		var listQuantity *string
		var listDone *bool
		if err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.categoryName, &listID, &listQuantity, &listDone); err != nil {
			return nil, err
		}

		if listID != nil {
			item.List = &ListItem{
				ID:       *listID,
				Quantity: *listQuantity,
				Done:     *listDone,
			}
		}

		ret = append(ret, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func GetItem(ctx context.Context, id int) (Item, error) {
	ret := Item{}
	var listID *int
	err := db.QueryRowContext(ctx, `
SELECT item.id, item.category_id, item.name, category.name AS category_name, item_list.id AS list_id
FROM item
LEFT JOIN category ON (item.category_id = category.id)
LEFT JOIN item_list ON (item_list.item_id = item.id)
WHERE item.id = $1`, id).
		Scan(&ret.ID, &ret.CategoryID, &ret.Name, &ret.categoryName, &listID)
	if err != nil {
		return ret, err
	}

	if listID != nil {
		ret.List, err = GetListItem(ctx, *listID)
	}

	return ret, err
}

func GetItemByName(ctx context.Context, name string) (Item, error) {
	ret := Item{}
	var inList *bool
	err := db.QueryRowContext(ctx, `
SELECT item.id, item.category_id, item.name, category.name AS category_name,
	(SELECT TRUE FROM item_list WHERE item_list.item_id = item.id) AS in_list
FROM item
LEFT JOIN category ON (item.category_id = category.id)
WHERE item.name = $1`, name).
		Scan(&ret.ID, &ret.CategoryID, &ret.Name, &ret.categoryName, &inList)

	if inList != nil {
		ret.List = &ListItem{}
	}

	return ret, err
}

func ItemChangeCategory(ctx context.Context, id int, categoryID int) error {
	_, err := db.ExecContext(ctx, `UPDATE item SET category_id = $2 WHERE id = $1`, id, categoryID)
	return err
}

func AddItem(ctx context.Context, i Item) error {
	if err := i.Validate(ctx); err != nil {
		return fmt.Errorf("invalid item: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if i.ID == 0 {
		_, err = insertWithID(ctx, tx,
			`INSERT INTO item (category_id, name) VALUES ($1, $2) RETURNING id`,
			i.CategoryID,
			i.Name,
		)
		if err != nil {
			return errors.Join(tx.Rollback(), err)
		}
	}

	return tx.Commit()
}

func EditItem(ctx context.Context, i Item) error {
	if err := i.Validate(ctx); err != nil {
		return fmt.Errorf("invalid item: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
UPDATE item SET
	category_id = $2,
	name = $3
WHERE id = $1`, i.ID, i.CategoryID, i.Name)
	if err != nil {
		return errors.Join(tx.Rollback(), err)
	}

	if i.List != nil {
		_, err := db.ExecContext(ctx, `
UPDATE item_list SET
	quantity = $2
WHERE id = $1`, i.List.ID, i.List.Quantity)
		if err != nil {
			return errors.Join(tx.Rollback(), err)
		}
	}

	return tx.Commit()
}

func DeleteItem(ctx context.Context, id int) error {
	_, err := db.ExecContext(ctx, `DELETE FROM item WHERE id = $1`, id)
	if err != nil {
		// Check if the error is due to foreign key constraint (item in recipe)
		if err.Error() != "" {
			// Try to get recipes using this item
			recipes, recipeErr := GetRecipesUsingItem(ctx, id)
			if recipeErr == nil && len(recipes) > 0 {
				recipeNames := strings.Builder{}
				for i, recipe := range recipes {
					if i > 0 {
						recipeNames.WriteString(", ")
					}
					recipeNames.WriteString(recipe.Name)
				}
				return fmt.Errorf("cannot delete item: it is used in recipes: %s", recipeNames.String())
			}
		}
	}
	return err
}
