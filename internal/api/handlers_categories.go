package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) categoriesListHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, categories)
}

func (s *Server) categoriesGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	category, err := models.GetCategory(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "category")
		} else {
			internalError(w, err)
		}
		return
	}

	items, err := category.Items(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	type response struct {
		models.Category
		Items []models.Item `json:"items"`
	}

	writeJSON(w, http.StatusOK, response{
		Category: category,
		Items:    items,
	})
}

func (s *Server) categoriesCreateHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		StoreID     int    `json:"store_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	if body.Name == "" {
		badRequest(w, "name is required")
		return
	}
	if body.StoreID == 0 {
		badRequest(w, "store_id is required")
		return
	}

	cat := models.Category{
		StoreID:     body.StoreID,
		Name:        body.Name,
		Description: body.Description,
	}

	if err := models.AddCategory(r.Context(), cat); err != nil {
		internalError(w, err)
		return
	}

	// Reload to get the generated ID and item_count
	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	for _, c := range categories {
		if c.Name == body.Name && c.StoreID == body.StoreID {
			writeJSON(w, http.StatusCreated, c)
			return
		}
	}

	// Fallback: return the struct we built (ID will be zero)
	writeJSON(w, http.StatusCreated, cat)
}

func (s *Server) categoriesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	existing, err := models.GetCategory(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "category")
		} else {
			internalError(w, err)
		}
		return
	}

	var body struct {
		StoreID     int    `json:"store_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	if body.Name == "" {
		badRequest(w, "name is required")
		return
	}
	if body.StoreID == 0 {
		badRequest(w, "store_id is required")
		return
	}

	existing.StoreID = body.StoreID
	existing.Name = body.Name
	existing.Description = body.Description

	if err := models.EditCategory(r.Context(), existing); err != nil {
		internalError(w, err)
		return
	}

	updated, err := models.GetCategory(r.Context(), id)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) categoriesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	if _, err := models.GetCategory(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "category")
		} else {
			internalError(w, err)
		}
		return
	}

	if err := models.DeleteCategory(r.Context(), id); err != nil {
		if err.Error() == "category is still in use" {
			conflict(w, err.Error())
		} else {
			internalError(w, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
