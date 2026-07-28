package api

import (
	"encoding/json"
	"net/http"

	"github.com/taiidani/groceries/internal/db/models"
)

func (s *Server) storesListHandler(w http.ResponseWriter, r *http.Request) {
	stores, err := s.svc.ListStores(r.Context())
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

	detail, err := s.svc.GetStore(r.Context(), id)
	if err != nil {
		listServiceError(w, err, "store")
		return
	}

	type response struct {
		models.Store
		Categories []models.Category `json:"categories"`
	}

	writeJSON(w, http.StatusOK, response{
		Store:      detail.Store,
		Categories: detail.Categories,
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

	store, err := s.svc.CreateStore(r.Context(), req.Name)
	if err != nil {
		listServiceError(w, err, "store")
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

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	store, err := s.svc.UpdateStore(r.Context(), id, req.Name)
	if err != nil {
		listServiceError(w, err, "store")
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

	if err := s.svc.DeleteStore(r.Context(), id); err != nil {
		listServiceError(w, err, "store")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
