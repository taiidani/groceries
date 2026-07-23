package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/taiidani/groceries/internal/db/models"
)

func (s *Server) storesListHandler(w http.ResponseWriter, r *http.Request) {
	stores, err := s.db.ListStores(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stores)
}

func (s *Server) storesGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	store, err := s.db.GetStore(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "store")
		} else {
			internalError(w, err)
		}
		return
	}

	categories, err := s.db.ListCategoriesForStore(r.Context(), store.ID)
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

	if err := s.db.ValidateStore(r.Context(), models.Store{Name: req.Name}); err != nil {
		badRequest(w, err.Error())
		return
	}

	store, err := s.db.CreateStore(r.Context(), req.Name)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, store)
}

func (s *Server) storesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	existing, err := s.db.GetStore(r.Context(), id)
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

	if err := s.db.ValidateStore(r.Context(), models.Store{ID: id, Name: req.Name}); err != nil {
		badRequest(w, err.Error())
		return
	}

	existing.Name = req.Name
	store, err := s.db.UpdateStore(r.Context(), models.UpdateStoreParams{
		ID:   id,
		Name: req.Name,
	})
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, store)
}

func (s *Server) storesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	if _, err := s.db.GetStore(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "store")
		} else {
			internalError(w, err)
		}
		return
	}

	if err := s.db.DeleteStore(r.Context(), id); err != nil {
		// DeleteStore returns a descriptive error when the store is in use
		conflict(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
