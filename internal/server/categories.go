package server

import (
	"fmt"
	"net/http"

	"github.com/taiidani/groceries/internal/models"
)

type categoriesBag struct {
	baseBag
	Categories []models.Category
}

func (s *Server) categoriesHandler(w http.ResponseWriter, r *http.Request) {
	bag := categoriesBag{baseBag: s.newBag(r.Context())}

	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Categories = categories

	template := "categories.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) categoryAddHandler(w http.ResponseWriter, r *http.Request) {
	newCategory := models.Category{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}

	// Validate inputs
	if len(newCategory.Name) < 3 {
		errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("provided name needs to be at least 3 characters"))
		return
	}

	// Check for existing category
	existingCategories, err := models.LoadCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("could not load categories"))
		return
	}
	for _, cat := range existingCategories {
		if cat.ID == newCategory.ID {
			errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("category already found"))
			return
		}
	}

	// Add the new category
	err = models.AddCategory(r.Context(), newCategory)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventCategory, nil)

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (s *Server) categoryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteCategory(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventCategory, nil)

	http.Redirect(w, r, "/categories", http.StatusFound)
}
