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

	// Broadcast the change
	s.sseServer.announce(sseEventBag)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) bagUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("could not parse id: %w", err))
		return
	}

	// Update the category if it has changed
	category := r.FormValue("category")
	if category != "" {
		categoryID, err := strconv.Atoi(r.FormValue("category"))
		if err != nil {
			errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("could not parse category: %w", err))
			return
		}

		err = models.ItemChangeCategory(r.Context(), id, categoryID)
		if err != nil {
			errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("unable to update category: %w", err))
			return
		}
	}

	// Update the name, if it has changed
	fullName := r.FormValue("name")
	if fullName != "" {
		// Parse the name (quantity) into a name, quantity pair
		name, quantity, err := parseItemName(fullName)
		if err != nil {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}

		err = models.BagUpdateItemName(r.Context(), id, name, quantity)
		if err != nil {
			errorResponse(r.Context(), w, http.StatusInternalServerError, err)
			return
		}
	}

	// Broadcast the change
	s.sseServer.announce(sseEventBag)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) bagDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteFromBag(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, fmt.Errorf("could not remove item from bag: %w", err))
		return
	}

	// Broadcast the change
	s.sseServer.announce(sseEventBag)

	http.Redirect(w, r, "/", http.StatusFound)
}

// bagDoneHandler will clear the bag by setting all items' `InBag` to false
func (s *Server) bagDoneHandler(w http.ResponseWriter, r *http.Request) {
	err := models.SaveBag(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.announce(sseEventBag, sseEventList)

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
