package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

type itemsBag struct {
	baseBag
	Categories     []models.Category
	ListCategories []categoryWithItems
}

func (s *Server) itemsHandler(w http.ResponseWriter, r *http.Request) {
	bag := itemsBag{baseBag: s.newBag(r.Context())}

	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
	bag.Categories = categories

	items, err := models.LoadItems(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	for _, cat := range categories {
		add := []models.Item{}
		for _, item := range items {
			if item.CategoryID == cat.ID {
				add = append(add, item)
			}
		}

		if len(add) > 0 {
			bag.ListCategories = append(bag.ListCategories, categoryWithItems{
				Category: models.Category{
					ID:          cat.ID,
					Description: cat.Description,
					Name:        cat.Name,
				},
				Items: add,
			})
		}
	}

	template := "items.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) itemAddHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
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
		errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("provided category not found"))
		return
	}

	newItem := models.Item{
		CategoryID: categoryID,
		Name:       r.FormValue("name"),
	}

	err = models.AddItem(r.Context(), newItem)
	if err != nil {
		err = fmt.Errorf("could not add item: %w", err)
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.announce(sseEventList)

	http.Redirect(w, r, "/items", http.StatusFound)
}

func (s *Server) itemListAddHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	err = models.ListAddItem(r.Context(), id, "")
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.announce(sseEventList)

	http.Redirect(w, r, "/items", http.StatusFound)
}

func (s *Server) itemListDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteFromList(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.announce(sseEventList)

	http.Redirect(w, r, "/items", http.StatusFound)
}

func (s *Server) itemDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteItem(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.announce(sseEventList)

	http.Redirect(w, r, "/items", http.StatusFound)
}
