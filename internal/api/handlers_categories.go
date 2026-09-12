package api

import (
	"net/http"

	"github.com/taiidani/groceries/internal/db/models"
)

func (s *Server) categoriesListHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := s.svc.ListCategoriesWithItemCount(r.Context())
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
