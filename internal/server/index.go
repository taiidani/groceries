package server

import (
	"net/http"
	"sort"

	"github.com/taiidani/groceries/internal/models"
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

	items, err := models.LoadItems(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	for _, item := range items {
		if item.List == nil {
			bag.Items = append(bag.Items, itemWithCategory{
				Category: item.CategoryName(),
				Name:     item.Name,
			})
		}
	}

	sort.Slice(bag.Items, func(i int, j int) bool {
		return bag.Items[i].Name < bag.Items[j].Name
	})

	template := "index.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) indexListHandler(w http.ResponseWriter, r *http.Request) {
	type indexListBag struct {
		baseBag
		List []storeWithCategories
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

	renderHtml(w, http.StatusOK, "index_list.gohtml", bag)
}

func (s *Server) indexCartHandler(w http.ResponseWriter, r *http.Request) {
	type indexCartBag struct {
		baseBag
		Total          int
		TotalDone      int
		DoneCategories []categoryWithItems
	}

	bag := indexCartBag{baseBag: s.newBag(r.Context())}

	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	listItems, err := models.LoadList(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	for _, cat := range categories {
		addDone := []models.Item{}
		for _, item := range listItems {
			if item.CategoryID != cat.ID {
				continue
			} else if item.List.Done {
				bag.Total++
				bag.TotalDone++
				addDone = append(addDone, item)
			} else {
				bag.Total++
			}
		}

		if len(addDone) > 0 {
			bag.DoneCategories = append(bag.DoneCategories, categoryWithItems{
				Category: models.Category{
					ID:          cat.ID,
					Description: cat.Description,
					Name:        cat.Name,
				},
				Items: addDone,
			})
		}
	}

	renderHtml(w, http.StatusOK, "index_cart.gohtml", bag)
}
