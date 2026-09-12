package api

import (
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
