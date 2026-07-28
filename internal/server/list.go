package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/service"
)

func (s *Server) listAddHandler(w http.ResponseWriter, r *http.Request) {
	var itemID *int32
	name := r.FormValue("name")

	if name == "" && r.PathValue("id") != "" {
		id, convErr := strconv.Atoi(r.PathValue("id"))
		if convErr != nil {
			errorResponse(w, r, http.StatusBadRequest, convErr)
			return
		}
		id32 := int32(id)
		itemID = &id32
	}

	entry, err := s.svc.AddItem(r.Context(), itemID, name, r.FormValue("quantity"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			errorResponse(w, r, http.StatusBadRequest, err)
		case errors.Is(err, service.ErrNotFound):
			errorResponse(w, r, http.StatusInternalServerError, err)
		case errors.Is(err, service.ErrConflict):
			errorResponse(w, r, http.StatusConflict, err)
		default:
			errorResponse(w, r, http.StatusInternalServerError, fmt.Errorf("unable to add item: %w", err))
		}
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = fmt.Sprintf("/item/%d", entry.ItemID)
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) listDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := s.svc.RemoveItem(r.Context(), int32(id)); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/item/" + r.PathValue("id")
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) listDoneHandler(w http.ResponseWriter, r *http.Request) {
	s.listMarkDone(w, r, true)
}

func (s *Server) listUnDoneHandler(w http.ResponseWriter, r *http.Request) {
	s.listMarkDone(w, r, false)
}

func (s *Server) listMarkDone(w http.ResponseWriter, r *http.Request, done bool) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := s.svc.MarkDone(r.Context(), int32(id), done); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) finishHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Finish(r.Context()); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
