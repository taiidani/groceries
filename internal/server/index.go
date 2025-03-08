package server

import (
	"net/http"

	"github.com/taiidani/groceries/internal/models"
)

type indexBag struct {
	baseBag
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

	template := "index.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}
