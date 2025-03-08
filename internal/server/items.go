package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/taiidani/groceries/internal/data"
	"github.com/taiidani/groceries/internal/models"
)

const listDefaultExpiration = time.Hour * 8760

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

	// Parse the name (quantity) into a name, quantity pair
	name, quantity, err := parseItemName(r.FormValue("name"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	newItem := models.Item{
		ID:         base64.StdEncoding.EncodeToString([]byte(r.FormValue("name"))),
		CategoryID: categoryID,
		Name:       name,
		InBag:      r.FormValue("in-bag") == "true",
		Quantity:   quantity,
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
	err = s.backend.Set(r.Context(), models.ListDBKey, list, listDefaultExpiration)
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
	err = s.backend.Set(r.Context(), models.ListDBKey, list, listDefaultExpiration)
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
	err = s.backend.Set(r.Context(), models.ListDBKey, list, listDefaultExpiration)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// bagDoneHandler will clear the bag by setting all items' `InBag` to false
func (s *Server) bagDoneHandler(w http.ResponseWriter, r *http.Request) {
	var list models.List
	err := s.backend.Get(r.Context(), models.ListDBKey, &list)
	if err != nil {
		if !errors.Is(err, data.ErrKeyNotFound) {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}

	for i, cat := range list.Categories {
		for j := range cat.Items {
			list.Categories[i].Items[j].InBag = false
		}
	}

	// Save the updated list
	err = s.backend.Set(r.Context(), models.ListDBKey, list, listDefaultExpiration)
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

	err = s.backend.Set(r.Context(), models.ListDBKey, list, listDefaultExpiration)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func parseItemName(name string) (string, string, error) {
	quantity := ""

	// Determine the quantity, if present
	re := regexp.MustCompile(`^(.+) \((.+)\)$`)
	matches := re.FindStringSubmatch(name)
	if len(matches) == 3 {
		name = matches[1]
		quantity = matches[2]
	}

	if len(name) < 3 {
		return name, quantity, errors.New("provided name needs to be at least 3 characters")
	}

	return name, quantity, nil
}
