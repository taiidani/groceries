package server

import (
	"context"

	"github.com/taiidani/groceries/internal/models"
)

type storeHierarchyInput struct {
	ExcludeEmptyGroupings bool
	ExcludeDoneItems      bool
	OnlyListItems         bool
}

func loadStoreHierarchy(ctx context.Context, input storeHierarchyInput) ([]storeWithCategories, error) {
	ret := []storeWithCategories{}

	stores, err := models.LoadStores(ctx)
	if err != nil {
		return ret, err
	}

	categories, err := models.LoadCategories(ctx)
	if err != nil {
		return ret, err
	}

	var items []models.Item
	if input.OnlyListItems {
		items, err = models.LoadList(ctx)
	} else {
		items, err = models.LoadItems(ctx)
	}
	if err != nil {
		return ret, err
	}

	for _, store := range stores {
		addStore := storeWithCategories{Store: store}

		for _, cat := range categories {
			if cat.StoreID != addStore.Store.ID {
				continue
			}

			addItem := []models.Item{}
			for _, item := range items {
				if item.CategoryID != cat.ID {
					continue
				}
				if input.ExcludeDoneItems && item.List != nil && item.List.Done {
					continue
				}

				addItem = append(addItem, item)
			}

			if !input.ExcludeEmptyGroupings || len(addItem) > 0 {
				addStore.Categories = append(addStore.Categories, categoryWithItems{
					Category: cat,
					Items:    addItem,
				})
			}
		}

		if !input.ExcludeEmptyGroupings || len(addStore.Categories) > 0 {
			ret = append(ret, addStore)
		}
	}

	return ret, nil
}
