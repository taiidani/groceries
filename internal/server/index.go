package server

import (
	"net/http"
	"sort"

	"github.com/taiidani/groceries/internal/client"
)

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	type itemWithCategory struct {
		Category string
		Name     string
	}

	type indexBag struct {
		baseBag
		Items []itemWithCategory
	}

	bag := indexBag{baseBag: s.newBag(r.Context())}

	apiClient := clientFromContext(r.Context())

	items, err := apiClient.ListItems(r.Context(), nil)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	categories, err := apiClient.ListCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	catNames := make(map[int]string, len(categories))
	for _, cat := range categories {
		catNames[cat.ID] = cat.Name
	}

	for _, item := range items {
		if item.List == nil {
			bag.Items = append(bag.Items, itemWithCategory{
				Category: catNames[item.CategoryID],
				Name:     item.Name,
			})
		}
	}

	sort.Slice(bag.Items, func(i, j int) bool {
		return bag.Items[i].Name < bag.Items[j].Name
	})

	renderHtml(w, http.StatusOK, "index.gohtml", bag)
}

func (s *Server) indexListHandler(w http.ResponseWriter, r *http.Request) {
	type indexListBag struct {
		baseBag
		Total     int
		TotalDone int
		List      []storeWithCategories
	}

	bag := indexListBag{baseBag: s.newBag(r.Context())}

	var err error
	bag.List, err = loadStoreHierarchy(r.Context(), storeHierarchyInput{
		OnlyListItems:         true,
		ExcludeDoneItems:      true,
		ExcludeEmptyGroupings: true,
	})
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	apiClient := clientFromContext(r.Context())
	listItems, err := apiClient.ListShoppingList(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	for _, item := range listItems {
		bag.Total++
		if item.List != nil && item.List.Done {
			bag.TotalDone++
		}
	}

	renderHtml(w, http.StatusOK, "index_list.gohtml", bag)
}

func (s *Server) indexCartHandler(w http.ResponseWriter, r *http.Request) {
	type indexCartBag struct {
		baseBag
		DoneCategories []categoryWithItems
	}

	bag := indexCartBag{baseBag: s.newBag(r.Context())}

	apiClient := clientFromContext(r.Context())

	categories, err := apiClient.ListCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	listItems, err := apiClient.ListShoppingList(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	for _, cat := range categories {
		var done []client.Item
		for _, item := range listItems {
			if item.CategoryID == cat.ID && item.List != nil && item.List.Done {
				done = append(done, item)
			}
		}

		if len(done) > 0 {
			bag.DoneCategories = append(bag.DoneCategories, categoryWithItems{
				Category: cat,
				Items:    done,
			})
		}
	}

	renderHtml(w, http.StatusOK, "index_cart.gohtml", bag)
}
