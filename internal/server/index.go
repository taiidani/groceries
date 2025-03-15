package server

import (
	"net/http"

	"github.com/taiidani/groceries/internal/models"
)

type indexBag struct {
	baseBag
	Total          int
	TotalDone      int
	Categories     []models.Category
	BagItems       []models.Item
	ListCategories []categoryWithItems
	DoneCategories []categoryWithItems
}

type categoryWithItems struct {
	models.Category
	Items []models.Item
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	bag := indexBag{baseBag: s.newBag(r)}

	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}
	bag.Categories = categories

	bagItems, err := models.LoadBag(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}
	bag.BagItems = bagItems

	listItems, err := models.LoadList(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	for _, cat := range categories {
		addList := []models.Item{}
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

	template := "index.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}
