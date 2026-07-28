package service

import (
	"context"
	"fmt"

	"github.com/taiidani/groceries/internal/db/models"
)

// HierarchyItem is an item as presented in the store hierarchy, carrying its
// shopping list state. List is nil when the item is not on the shopping list.
type HierarchyItem struct {
	ID         int32
	CategoryID int32
	Name       string
	List       *HierarchyListEntry
}

// HierarchyListEntry holds the list-specific state for a hierarchy item.
type HierarchyListEntry struct {
	ID       int32
	Quantity string
	Done     bool
}

// CategoryWithItems is a category together with the items that belong to it.
type CategoryWithItems struct {
	models.Category
	Items []HierarchyItem
}

// StoreWithCategories is a store together with its categories, each carrying
// their items.
type StoreWithCategories struct {
	models.Store
	Categories []CategoryWithItems
}

// HierarchyInput controls how LoadStoreHierarchy filters its results.
type HierarchyInput struct {
	// ExcludeEmptyGroupings drops categories with no items and stores with no
	// categories from the result.
	ExcludeEmptyGroupings bool
	// ExcludeDoneItems drops items whose list entry is marked done.
	ExcludeDoneItems bool
	// OnlyListItems restricts the aggregation to items currently on the
	// shopping list.
	OnlyListItems bool
}

// LoadStoreHierarchy aggregates stores -> categories -> items for display
// purposes, composing the existing list queries.
func (s *Service) LoadStoreHierarchy(ctx context.Context, input HierarchyInput) ([]StoreWithCategories, error) {
	stores, err := s.queries.ListStores(ctx)
	if err != nil {
		return nil, fmt.Errorf("list stores: %w", err)
	}

	categories, err := s.queries.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	itemRows, err := s.queries.SummarizeItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("summarize items: %w", err)
	}

	items := make([]HierarchyItem, 0, len(itemRows))
	for _, row := range itemRows {
		if input.OnlyListItems && !row.ListID.Valid {
			continue
		}
		if input.ExcludeDoneItems && row.ListID.Valid && row.ListDone.Bool {
			continue
		}

		item := HierarchyItem{
			ID:         row.ID,
			CategoryID: row.CategoryID,
			Name:       row.Name,
		}
		if row.ListID.Valid {
			item.List = &HierarchyListEntry{
				ID:       row.ListID.Int32,
				Quantity: row.ListQuantity.String,
				Done:     row.ListDone.Bool,
			}
		}
		items = append(items, item)
	}

	ret := []StoreWithCategories{}
	for _, store := range stores {
		addStore := StoreWithCategories{Store: store}

		for _, cat := range categories {
			if cat.StoreID != store.ID {
				continue
			}

			var addItems []HierarchyItem
			for _, item := range items {
				if item.CategoryID == cat.ID {
					addItems = append(addItems, item)
				}
			}

			if !input.ExcludeEmptyGroupings || len(addItems) > 0 {
				addStore.Categories = append(addStore.Categories, CategoryWithItems{
					Category: cat,
					Items:    addItems,
				})
			}
		}

		if !input.ExcludeEmptyGroupings || len(addStore.Categories) > 0 {
			ret = append(ret, addStore)
		}
	}

	return ret, nil
}
