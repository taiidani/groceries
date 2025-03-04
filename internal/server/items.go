package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/taiidani/groceries/internal/data"
	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) itemAddHandler(w http.ResponseWriter, r *http.Request) {
	var list models.List
	err := s.backend.Get(r.Context(), models.ListDBKey, &list)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}

	categoryID := r.FormValue("category")
	var category *models.Category
	for _, cat := range list.Categories {
		if cat.ID == categoryID {
			category = cat
		}
	}
	if category == nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("provided category not found"))
		return
	}

	newItem := models.Item{
		ID:       base64.StdEncoding.EncodeToString([]byte(r.FormValue("name"))),
		Name:     r.FormValue("name"),
		Quantity: r.FormValue("quantity"),
	}

	// Validate inputs
	if len(newItem.Name) < 3 {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("provided name needs to be at least 3 characters"))
		return
	}

	// Check for existing item
	for _, item := range category.Items {
		if item.ID == newItem.ID {
			errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("item already found"))
			return
		}
	}

	// Add the new item
	category.Items = append(category.Items, newItem)
	sort.Slice(category.Items, func(i, j int) bool {
		return category.Items[i].Name < category.Items[j].Name
	})

	// And save
	err = s.backend.Set(r.Context(), models.ListDBKey, list, time.Hour*8760)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) itemDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var list models.List
	err := s.backend.Get(r.Context(), models.ListDBKey, &list)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}

	for i, cat := range list.Categories {
		for j, item := range cat.Items {
			if item.ID == r.FormValue("id") {
				list.Categories[i].Items = append(list.Categories[i].Items[0:j], list.Categories[i].Items[j+1:]...)
			}
		}
	}
	err = s.backend.Set(r.Context(), models.ListDBKey, list, time.Hour*8760)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) itemDoneHandler(w http.ResponseWriter, r *http.Request) {
	var list models.List
	err := s.backend.Get(r.Context(), models.ListDBKey, &list)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}

	for i, cat := range list.Categories {
		for j, item := range cat.Items {
			if item.ID == r.FormValue("id") {
				list.Categories[i].Items[j].Done = !list.Categories[i].Items[j].Done
			}
		}
	}
	err = s.backend.Set(r.Context(), models.ListDBKey, list, time.Hour*8760)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) finishHandler(w http.ResponseWriter, r *http.Request) {
	var list models.List
	err := s.backend.Get(r.Context(), models.ListDBKey, &list)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}

	for i, cat := range list.Categories {
		newItems := []models.Item{}
		for _, item := range cat.Items {
			if !item.Done {
				newItems = append(newItems, item)
			}
		}
		list.Categories[i].Items = newItems
	}

	err = s.backend.Set(r.Context(), models.ListDBKey, list, time.Hour*8760)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
