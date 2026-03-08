package server

import (
	"context"

	"github.com/taiidani/groceries/internal/client"
)

type storeHierarchyInput struct {
	ExcludeEmptyGroupings bool
	ExcludeDoneItems      bool
	OnlyListItems         bool
}

func loadStoreHierarchy(ctx context.Context, input storeHierarchyInput) ([]storeWithCategories, error) {
	ret := []storeWithCategories{}

	apiClient := clientFromContext(ctx)

	stores, err := apiClient.ListStores(ctx)
	if err != nil {
		return ret, err
	}

	categories, err := apiClient.ListCategories(ctx)
	if err != nil {
		return ret, err
	}

	var items []client.Item
	if input.OnlyListItems {
		items, err = apiClient.ListShoppingList(ctx)
		if err != nil {
			return ret, err
		}
	} else {
		items, err = apiClient.ListItems(ctx, nil)
		if err != nil {
			return ret, err
		}
	}

	for _, store := range stores {
		addStore := storeWithCategories{Store: store}

		for _, cat := range categories {
			if cat.StoreID != addStore.Store.ID {
				continue
			}

			addItems := []client.Item{}
			for _, item := range items {
				if item.CategoryID != cat.ID {
					continue
				}
				if input.ExcludeDoneItems && item.List != nil && item.List.Done {
					continue
				}

				addItems = append(addItems, item)
			}

			if !input.ExcludeEmptyGroupings || len(addItems) > 0 {
				addStore.Categories = append(addStore.Categories, categoryWithItems{
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
