package server

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) listItemGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	type bag struct {
		ID       int
		Name     string
		Quantity string
	}
	item, err := models.GetListItem(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	data := bag{
		ID:       item.ID,
		Name:     item.Name,
		Quantity: item.Quantity,
	}

	template := "list_item.gohtml"
	renderHtml(w, http.StatusOK, template, data)
}

func (s *Server) listItemSaveHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	item, err := models.GetListItem(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
	item.Quantity = r.FormValue("quantity")

	// Validate inputs
	if err := item.Validate(r.Context()); err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	// Update the item
	err = models.EditListItemQuantity(r.Context(), item.ID, item.Quantity)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventList, nil)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) listAddHandler(w http.ResponseWriter, r *http.Request) {
	item, err := models.GetItemByName(r.Context(), r.FormValue("name"))
	switch {
	case err == nil:
	case errors.Is(err, sql.ErrNoRows):
		// The item doesn't exist yet. That's okay!
		// Let's create a new one
		newItem := models.Item{
			Name:       r.FormValue("name"),
			CategoryID: models.UncategorizedCategoryID,
		}
		err = models.AddItem(r.Context(), newItem)
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("unable to add item: %w", err))
			return
		}

		item, err = models.GetItemByName(r.Context(), r.FormValue("name"))
		if err != nil {
			errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("unable to retrieve added item: %w", err))
			return
		}
	default:
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	err = models.ListAddItem(r.Context(), item.ID, r.FormValue("quantity"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventList, nil)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) listDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteFromList(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventList, nil)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) listDoneHandler(w http.ResponseWriter, r *http.Request) {
	err := models.MarkItemDone(r.Context(), r.FormValue("id"), true)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventList, nil)
	s.sseServer.Publish(r.Context(), sseEventCart, nil)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) listUnDoneHandler(w http.ResponseWriter, r *http.Request) {
	err := models.MarkItemDone(r.Context(), r.FormValue("id"), false)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventList, nil)
	s.sseServer.Publish(r.Context(), sseEventCart, nil)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) finishHandler(w http.ResponseWriter, r *http.Request) {
	err := models.FinishShopping(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventCart, nil)

	http.Redirect(w, r, "/", http.StatusFound)
}
