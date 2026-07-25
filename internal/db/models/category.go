package models

import (
	"context"
	"errors"
	"fmt"
)

func (q *Queries) ValidateCategory(ctx context.Context, c Category) error {
	var vErr error

	if len(c.Name) < 3 {
		vErr = errors.Join(vErr, errors.New("provided name needs to be at least 3 characters"))
	}

	if _, err := q.GetStore(ctx, c.StoreID); err != nil {
		vErr = errors.Join(vErr, fmt.Errorf("store not found: %w", err))
	}

	// Check for existing category
	if c.ID == 0 {
		if existingCategories, err := q.ListCategories(ctx); err != nil {
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
