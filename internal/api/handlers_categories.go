package api

import (
	"encoding/json"
	"net/http"

	"github.com/taiidani/groceries/internal/db/models"
)

func (s *Server) categoriesListHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := s.svc.ListCategories(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, categories)
}

func (s *Server) categoriesGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	detail, err := s.svc.GetCategory(r.Context(), id)
	if err != nil {
		listServiceError(w, err, "category")
		return
	}

	type response struct {
		models.Category
		Items []models.Item `json:"items"`
	}

	writeJSON(w, http.StatusOK, response{
		Category: detail.Category,
		Items:    detail.Items,
	})
}

func (s *Server) categoriesCreateHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StoreID     int32  `json:"store_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	cat, err := s.svc.CreateCategory(r.Context(), body.StoreID, body.Name, body.Description)
	if err != nil {
		listServiceError(w, err, "category")
		return
	}

	writeJSON(w, http.StatusCreated, cat)
}

func (s *Server) categoriesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	var body struct {
		StoreID     int32  `json:"store_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	updated, err := s.svc.UpdateCategory(r.Context(), id, body.StoreID, body.Name, body.Description)
	if err != nil {
		listServiceError(w, err, "category")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) categoriesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	if err := s.svc.DeleteCategory(r.Context(), id); err != nil {
		listServiceError(w, err, "category")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
