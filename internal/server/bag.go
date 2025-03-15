package server

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) bagAddHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	categoryID := r.FormValue("category")
	var category *models.Category
	for i, cat := range categories {
		if cat.ID == categoryID {
			category = &categories[i]
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
		CategoryID: categoryID,
		Name:       name,
		Bag:        &models.BagItem{Quantity: quantity},
	}

	err = models.AddItem(r.Context(), newItem)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) bagUpdateHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	categoryID := r.FormValue("category")
	var category *models.Category
	for i, cat := range categories {
		if cat.ID == categoryID {
			category = &categories[i]
		}
	}
	if category == nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("provided category not found"))
		return
	}

	if r.FormValue("action") == "delete" {
		err := models.DeleteFromBag(r.Context(), r.FormValue("id"))
		if err != nil {
			errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("could not remove item from bag: %w", err))
			return
		}
	} else {
		// Parse the name (quantity) into a name, quantity pair
		name, quantity, err := parseItemName(r.FormValue("name"))
		if err != nil {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}

		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}

		newItem := models.Item{
			ID:         id,
			CategoryID: categoryID,
			Name:       name,
			Bag:        &models.BagItem{Quantity: quantity},
		}

		err = models.UpdateItem(r.Context(), newItem)
		if err != nil {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// bagDoneHandler will clear the bag by setting all items' `InBag` to false
func (s *Server) bagDoneHandler(w http.ResponseWriter, r *http.Request) {
	err := models.SaveBag(r.Context())
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
