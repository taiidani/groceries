package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) listGetHandler(w http.ResponseWriter, r *http.Request) {
	items, err := models.LoadList(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	total := len(items)
	totalDone := 0
	listItems := make([]listItemJSON, 0, len(items))

	for _, item := range items {
		if item.List == nil {
			continue
		}
		if item.List.Done {
			totalDone++
		}
		listItems = append(listItems, listItemToJSON(item))
	}

	type response struct {
		Items     []listItemJSON `json:"items"`
		Total     int            `json:"total"`
		TotalDone int            `json:"total_done"`
	}

	writeJSON(w, http.StatusOK, response{
		Items:     listItems,
		Total:     total,
		TotalDone: totalDone,
	})
}

func (s *Server) listAddItemHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ItemID   *int   `json:"item_id"`
		Name     string `json:"name"`
		Quantity string `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	var item models.Item

	switch {
	case req.ItemID != nil:
		var err error
		item, err = models.GetItem(r.Context(), *req.ItemID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				notFound(w, "item")
			} else {
				internalError(w, err)
			}
			return
		}

	case req.Name != "":
		var err error
		item, err = models.GetItemByName(r.Context(), req.Name)
		if errors.Is(err, sql.ErrNoRows) {
			// Create a new uncategorized item on the fly
			newItem := models.Item{
				Name:       req.Name,
				CategoryID: models.UncategorizedCategoryID,
			}
			if addErr := models.AddItem(r.Context(), newItem); addErr != nil {
				internalError(w, addErr)
				return
			}
			item, err = models.GetItemByName(r.Context(), req.Name)
		}
		if err != nil {
			internalError(w, err)
			return
		}

	default:
		badRequest(w, "one of item_id or name is required")
		return
	}

	if err := models.ListAddItem(r.Context(), item.ID, req.Quantity); err != nil {
		// Unique constraint violation means item is already on the list
		conflict(w, "item is already on the list")
		return
	}

	// Re-fetch the item so the response includes the populated list field
	updated, err := models.GetItem(r.Context(), item.ID)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, listItemToJSON(updated))
}

func (s *Server) listUpdateItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	var req struct {
		Quantity *string `json:"quantity"`
		Done     *bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	item, err := models.GetItem(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, "list item")
		} else {
			internalError(w, err)
		}
		return
	}

	if item.List == nil {
		notFound(w, "list item")
		return
	}

	if req.Quantity != nil {
		item.List.Quantity = *req.Quantity
	}

	if req.Done != nil {
		if err := models.MarkItemDone(r.Context(), strconv.Itoa(id), *req.Done); err != nil {
			internalError(w, err)
			return
		}
	}

	if req.Quantity != nil {
		if err := models.EditItem(r.Context(), item); err != nil {
			internalError(w, err)
			return
		}
	}

	updated, err := models.GetItem(r.Context(), id)
	if err != nil {
		internalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, listItemToJSON(updated))
}

func (s *Server) listRemoveItemHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		badRequest(w, "id is required")
		return
	}

	if _, err := strconv.Atoi(id); err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	if err := models.DeleteFromList(r.Context(), id); err != nil {
		internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listFinishHandler(w http.ResponseWriter, r *http.Request) {
	if err := models.FinishShopping(r.Context()); err != nil {
		internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// JSON representation helpers
// ---------------------------------------------------------------------------

type listItemJSON struct {
	ID         int    `json:"id"`
	ItemID     int    `json:"item_id"`
	ItemName   string `json:"item_name"`
	CategoryID int    `json:"category_id"`
	Quantity   string `json:"quantity"`
	Done       bool   `json:"done"`
}

func listItemToJSON(item models.Item) listItemJSON {
	out := listItemJSON{
		ItemID:     item.ID,
		ItemName:   item.Name,
		CategoryID: item.CategoryID,
	}
	if item.List != nil {
		out.ID = item.List.ID
		out.Quantity = item.List.Quantity
		out.Done = item.List.Done
	}
	return out
}
