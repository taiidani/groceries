package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) itemsListHandler(w http.ResponseWriter, r *http.Request) {
	items, err := models.LoadItems(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	// Apply query filters
	q := r.URL.Query()

	if rawID := q.Get("category_id"); rawID != "" {
		categoryID, err := strconv.Atoi(rawID)
		if err != nil {
			badRequest(w, "category_id must be an integer")
			return
		}
		filtered := items[:0]
		for _, item := range items {
			if item.CategoryID == categoryID {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	if rawInList := q.Get("in_list"); rawInList != "" {
		inList, err := strconv.ParseBool(rawInList)
		if err != nil {
			badRequest(w, "in_list must be a boolean")
			return
		}
		filtered := items[:0]
		for _, item := range items {
			if inList && item.List != nil {
				filtered = append(filtered, item)
			} else if !inList && item.List == nil {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	writeJSON(w, http.StatusOK, items)
}

func (s *Server) itemsCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CategoryID int    `json:"category_id"`
		Name       string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}
	if req.Name == "" {
		badRequest(w, "name is required")
		return
	}
	if req.CategoryID == 0 {
		badRequest(w, "category_id is required")
		return
	}

	newItem := models.Item{
		CategoryID: req.CategoryID,
		Name:       req.Name,
	}

	if err := models.AddItem(r.Context(), newItem); err != nil {
		internalError(w, err)
		return
	}

	created, err := models.GetItemByName(r.Context(), req.Name)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) itemsGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	item, err := models.GetItem(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "item")
		} else {
			internalError(w, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (s *Server) itemsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	item, err := models.GetItem(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "item")
		} else {
			internalError(w, err)
		}
		return
	}

	var req struct {
		CategoryID int    `json:"category_id"`
		Name       string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}
	if req.Name == "" {
		badRequest(w, "name is required")
		return
	}
	if req.CategoryID == 0 {
		badRequest(w, "category_id is required")
		return
	}

	item.Name = req.Name
	item.CategoryID = req.CategoryID

	if err := models.EditItem(r.Context(), item); err != nil {
		internalError(w, err)
		return
	}

	updated, err := models.GetItem(r.Context(), id)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) itemsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	_, err = models.GetItem(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "item")
		} else {
			internalError(w, err)
		}
		return
	}

	if err := models.DeleteItem(r.Context(), id); err != nil {
		// DeleteItem returns a descriptive error when the item is in use by recipes
		conflict(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
