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
	ListCategories []models.Category
	DoneCategories []models.Category
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	bag := indexBag{baseBag: s.newBag(r)}

	list := models.NewList(s.db)
	categories, err := list.LoadCategories(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	bag.Categories = categories
	for _, cat := range categories {
		listItems := []models.Item{}
		doneItems := []models.Item{}
		for _, item := range cat.Items {
			if item.InBag {
				bag.BagItems = append(bag.BagItems, item)
			} else if item.Done {
				doneItems = append(doneItems, item)
			} else {
				listItems = append(listItems, item)
			}
		}

		if len(listItems) > 0 {
			bag.ListCategories = append(bag.ListCategories, models.Category{
				ID:          cat.ID,
				Description: cat.Description,
				Name:        cat.Name,
				Items:       listItems,
			})
		}

		if len(doneItems) > 0 {
			bag.DoneCategories = append(bag.DoneCategories, models.Category{
				ID:          cat.ID,
				Description: cat.Description,
				Name:        cat.Name,
				Items:       doneItems,
			})
		}
	}

	// Count the total & total done for the progress bar
	for _, cat := range categories {
		for _, item := range cat.Items {
			bag.Total++
			if item.Done {
				bag.TotalDone++
			}
		}
	}

	template := "index.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}
