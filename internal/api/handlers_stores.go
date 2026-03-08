package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) storesListHandler(w http.ResponseWriter, r *http.Request) {
	stores, err := models.LoadStores(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stores)
}

func (s *Server) storesGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	store, err := models.GetStore(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "store")
		} else {
			internalError(w, err)
		}
		return
	}

	categories, err := store.Categories(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	type response struct {
		models.Store
		Categories []models.Category `json:"categories"`
	}

	writeJSON(w, http.StatusOK, response{
		Store:      store,
		Categories: categories,
	})
}

func (s *Server) storesCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		badRequest(w, "name is required")
		return
	}

	newStore := models.Store{Name: req.Name}
	if err := models.AddStore(r.Context(), newStore); err != nil {
		internalError(w, err)
		return
	}

	created, err := models.LoadStores(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	for _, s := range created {
		if s.Name == req.Name {
			writeJSON(w, http.StatusCreated, s)
			return
		}
	}

	errorJSON(w, http.StatusInternalServerError, "store created but could not be retrieved")
}

func (s *Server) storesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	existing, err := models.GetStore(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "store")
		} else {
			internalError(w, err)
		}
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		badRequest(w, "name is required")
		return
	}

	existing.Name = req.Name
	if err := models.EditStore(r.Context(), existing); err != nil {
		internalError(w, err)
		return
	}

	updated, err := models.GetStore(r.Context(), id)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) storesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	if _, err := models.GetStore(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "store")
		} else {
			internalError(w, err)
		}
		return
	}

	if err := models.DeleteStore(r.Context(), id); err != nil {
		// DeleteStore returns a descriptive error when the store is in use
		conflict(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
