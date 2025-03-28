package server

import (
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) listAddHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	err = models.ListAddItem(r.Context(), id, r.FormValue("quantity"))
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
