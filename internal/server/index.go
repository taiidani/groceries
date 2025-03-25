package server

import (
	"net/http"

	"github.com/taiidani/groceries/internal/models"
)

type categoryWithItems struct {
	models.Category
	Items []models.Item
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	type indexBag struct {
		baseBag
	}

	bag := indexBag{baseBag: s.newBag(r.Context())}

	template := "index.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) indexBagHandler(w http.ResponseWriter, r *http.Request) {
	type indexBagBag struct {
		baseBag
		Categories map[string][]models.Item
	}

	bag := indexBagBag{baseBag: s.newBag(r.Context())}

	items, err := models.LoadItems(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Categories = map[string][]models.Item{}
	for _, item := range items {
		if item.List == nil {
			bag.Categories[item.CategoryName()] = append(bag.Categories[item.CategoryName()], item)
		}
	}

	renderHtml(w, http.StatusOK, "index_bag.gohtml", bag)
}

func (s *Server) indexListHandler(w http.ResponseWriter, r *http.Request) {
	type indexListBag struct {
		baseBag
		ListCategories []categoryWithItems
	}

	bag := indexListBag{baseBag: s.newBag(r.Context())}

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
		addList := []models.Item{}
		for _, item := range listItems {
			if item.CategoryID != cat.ID {
				continue
			} else if !item.List.Done {
				addList = append(addList, item)
			}
		}

		if len(addList) > 0 {
			bag.ListCategories = append(bag.ListCategories, categoryWithItems{
				Category: models.Category{
					ID:          cat.ID,
					Description: cat.Description,
					Name:        cat.Name,
				},
				Items: addList,
			})
		}
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
