package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/taiidani/groceries/internal/data"
	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) categoriesHandler(w http.ResponseWriter, r *http.Request) {
	bag := indexBag{baseBag: s.newBag(r)}

	var list models.List
	err := s.backend.Get(r.Context(), models.ListDBKey, &list)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}
	bag.List = list

	template := "categories.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) categoryAddHandler(w http.ResponseWriter, r *http.Request) {
	var list models.List
	err := s.backend.Get(r.Context(), models.ListDBKey, &list)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}

	newCategory := &models.Category{
		ID:   base64.StdEncoding.EncodeToString([]byte(r.FormValue("name"))),
		Name: r.FormValue("name"),
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
	list.Categories = append(list.Categories, newCategory)
	sort.Slice(list.Categories, func(i, j int) bool {
		return list.Categories[i].Name < list.Categories[j].Name
	})

	// And save
	err = s.backend.Set(r.Context(), models.ListDBKey, list, listDefaultExpiration)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (s *Server) categoryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var list models.List
	err := s.backend.Get(r.Context(), models.ListDBKey, &list)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}

	for i, cat := range list.Categories {
		if cat.ID == r.FormValue("id") {
			list.Categories = append(list.Categories[0:i], list.Categories[i+1:]...)
			break
		}
	}
	err = s.backend.Set(r.Context(), models.ListDBKey, list, listDefaultExpiration)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusFound)
}
