package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) recipesListHandler(w http.ResponseWriter, r *http.Request) {
	recipes, err := models.LoadRecipes(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, recipes)
}

func (s *Server) recipesGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	recipe, err := models.GetRecipe(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "recipe")
		} else {
			internalError(w, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, recipe)
}

func (s *Server) recipesCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}
	if req.Name == "" {
		badRequest(w, "name is required")
		return
	}

	recipe := models.Recipe{
		Name:        req.Name,
		Description: req.Description,
	}

	id, err := models.AddRecipe(r.Context(), recipe)
	if err != nil {
		internalError(w, err)
		return
	}

	created, err := models.GetRecipe(r.Context(), id)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) recipesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}
	if req.Name == "" {
		badRequest(w, "name is required")
		return
	}

	_, err = models.GetRecipe(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "recipe")
		} else {
			internalError(w, err)
		}
		return
	}

	recipe := models.Recipe{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := models.EditRecipe(r.Context(), recipe); err != nil {
		internalError(w, err)
		return
	}

	updated, err := models.GetRecipe(r.Context(), id)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) recipesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	_, err = models.GetRecipe(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "recipe")
		} else {
			internalError(w, err)
		}
		return
	}

	if err := models.DeleteRecipe(r.Context(), id); err != nil {
		internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) recipesAddItemHandler(w http.ResponseWriter, r *http.Request) {
	recipeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	var req struct {
		ItemID   int    `json:"item_id"`
		Quantity string `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}
	if req.ItemID == 0 {
		badRequest(w, "item_id is required")
		return
	}

	_, err = models.GetRecipe(r.Context(), recipeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "recipe")
		} else {
			internalError(w, err)
		}
		return
	}

	_, err = models.GetItem(r.Context(), req.ItemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "item")
		} else {
			internalError(w, err)
		}
		return
	}

	if err := models.AddRecipeItem(r.Context(), recipeID, req.ItemID, req.Quantity); err != nil {
		// A unique constraint violation means the item is already in the recipe.
		conflict(w, "item is already in this recipe")
		return
	}

	updated, err := models.GetRecipe(r.Context(), recipeID)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, updated)
}

func (s *Server) recipesRemoveItemHandler(w http.ResponseWriter, r *http.Request) {
	recipeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	itemID, err := strconv.Atoi(r.PathValue("itemId"))
	if err != nil {
		badRequest(w, "itemId must be an integer")
		return
	}

	_, err = models.GetRecipe(r.Context(), recipeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "recipe")
		} else {
			internalError(w, err)
		}
		return
	}

	if err := models.RemoveRecipeItem(r.Context(), recipeID, itemID); err != nil {
		internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) recipesAddToListHandler(w http.ResponseWriter, r *http.Request) {
	recipeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	// Optional body to scope which items to add.
	var req struct {
		ItemIDs []int `json:"item_ids"`
	}
	// Ignore decode errors - body is optional.
	_ = json.NewDecoder(r.Body).Decode(&req)

	recipe, err := models.GetRecipe(r.Context(), recipeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "recipe")
		} else {
			internalError(w, err)
		}
		return
	}

	// Build a set of requested IDs for fast lookup (empty set = all items).
	filter := make(map[int]struct{}, len(req.ItemIDs))
	for _, id := range req.ItemIDs {
		filter[id] = struct{}{}
	}

	for _, recipeItem := range recipe.Items {
		if len(filter) > 0 {
			if _, ok := filter[recipeItem.ItemID]; !ok {
				continue
			}
		}

		// Skip items already on the list.
		if recipeItem.InList {
			continue
		}

		if err := models.ListAddItem(r.Context(), recipeItem.ItemID, recipeItem.Quantity); err != nil {
			internalError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
