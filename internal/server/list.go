package server

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) listAddHandler(w http.ResponseWriter, r *http.Request) {
	var item models.Item
	var err error
	switch {
	case r.FormValue("name") != "":
		item, err = models.GetItemByName(r.Context(), r.FormValue("name"))
		if errors.Is(err, sql.ErrNoRows) {
			// The item doesn't exist yet. That's okay!
			// Let's create a new one
			item = models.Item{
				Name:       r.FormValue("name"),
				CategoryID: models.UncategorizedCategoryID,
			}
			err = models.AddItem(r.Context(), item)
			if err != nil {
				errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("unable to add item: %w", err))
				return
			}

			item, err = models.GetItemByName(r.Context(), r.FormValue("name"))
		}
	case r.PathValue("id") != "":
		id, convErr := strconv.Atoi(r.PathValue("id"))
		if convErr != nil {
			errorResponse(w, r, http.StatusBadRequest, convErr)
			return
		}

		item, err = models.GetItem(r.Context(), id)
	}

	if err != nil {
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

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = fmt.Sprintf("/item/%d", item.ID)
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) listDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteFromList(r.Context(), r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventList, nil)

	http.Redirect(w, r, "/item/"+r.PathValue("id"), http.StatusFound)
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
