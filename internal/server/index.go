package server

import (
	"net/http"

	"github.com/taiidani/groceries/internal/models"
)

type indexBag struct {
	baseBag
	Total      int
	TotalDone  int
	Categories []models.Category
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
