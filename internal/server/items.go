package server

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

type itemsBag struct {
	baseBag
	Categories     []models.Category
	ListCategories []categoryWithItems
	Item           models.Item
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

func (s *Server) itemHandler(w http.ResponseWriter, r *http.Request) {
	bag := struct {
		baseBag
		Redirect   string
		Categories []models.Category
		Item       models.Item
	}{baseBag: s.newBag(r.Context())}

	bag.Redirect = r.URL.Query().Get("redirect")

	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
	bag.Categories = categories

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Item, err = models.GetItem(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorResponse(w, r, http.StatusNotFound, errors.New("item not found"))
		} else {
			errorResponse(w, r, http.StatusInternalServerError, err)
		}

		return
	}

	template := "item_edit.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) itemAddHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	categoryID := r.FormValue("categoryID")
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
	s.sseServer.Publish(r.Context(), sseEventList, nil)

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/items"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) itemEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	item, err := models.GetItem(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	item.CategoryID = r.FormValue("categoryID")
	item.Name = r.FormValue("name")
	if item.List != nil {
		item.List.Quantity = r.FormValue("quantity")
	}

	// Validate inputs
	if err := item.Validate(r.Context()); err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	// Add the new item
	err = models.EditItem(r.Context(), item)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventList, nil)

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/items"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) itemDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteItem(r.Context(), r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventList, nil)

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/items"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}
