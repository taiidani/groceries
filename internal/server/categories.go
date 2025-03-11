package server

import (
	"fmt"
	"net/http"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) categoriesHandler(w http.ResponseWriter, r *http.Request) {
	bag := indexBag{baseBag: s.newBag(r)}

	list := models.NewList(s.db)
	categories, err := list.LoadCategories(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	bag.Categories = categories

	template := "categories.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) categoryAddHandler(w http.ResponseWriter, r *http.Request) {
	list := models.NewList(s.db)

	newCategory := models.Category{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}

	// Validate inputs
	if len(newCategory.Name) < 3 {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("provided name needs to be at least 3 characters"))
		return
	}

	// Check for existing category
	for _, cat := range list.Categories {
		if cat.ID == newCategory.ID {
			errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("category already found"))
			return
		}
	}

	// Add the new category
	err := list.AddCategory(r.Context(), newCategory)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (s *Server) categoryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	list := models.NewList(s.db)

	err := list.DeleteCategory(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusFound)
}
