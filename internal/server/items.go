package server

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) itemAddHandler(w http.ResponseWriter, r *http.Request) {
	list := models.NewList(s.db)
	categories, err := list.LoadCategories(r.Context())
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
		InBag:      r.FormValue("in-bag") == "true",
		Quantity:   quantity,
	}

	err = list.AddItem(r.Context(), newItem)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) itemDeleteHandler(w http.ResponseWriter, r *http.Request) {
	list := models.NewList(s.db)

	err := list.DeleteItem(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) itemDoneHandler(w http.ResponseWriter, r *http.Request) {
	list := models.NewList(s.db)

	err := list.MarkItemDone(r.Context(), r.FormValue("id"), true)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) itemUnDoneHandler(w http.ResponseWriter, r *http.Request) {
	list := models.NewList(s.db)

	err := list.MarkItemDone(r.Context(), r.FormValue("id"), false)
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// bagDoneHandler will clear the bag by setting all items' `InBag` to false
func (s *Server) bagDoneHandler(w http.ResponseWriter, r *http.Request) {
	list := models.NewList(s.db)

	err := list.SaveBag(r.Context())
	if err != nil {
		errorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) finishHandler(w http.ResponseWriter, r *http.Request) {
	list := models.NewList(s.db)

	err := list.FinishShopping(r.Context())
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
