package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/service"
)

func (s *Server) listGetHandler(w http.ResponseWriter, r *http.Request) {
	summary, err := s.svc.GetList(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	listItems := make([]listItemJSON, 0, len(summary.Items))
	for _, entry := range summary.Items {
		listItems = append(listItems, listItemJSONFromEntry(entry))
	}

	type response struct {
		Items     []listItemJSON `json:"items"`
		Total     int            `json:"total"`
		TotalDone int            `json:"total_done"`
	}

	writeJSON(w, http.StatusOK, response{
		Items:     listItems,
		Total:     summary.Total,
		TotalDone: summary.TotalDone,
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

	var itemID *int32
	if req.ItemID != nil {
		id := int32(*req.ItemID)
		itemID = &id
	}

	entry, err := s.svc.AddItem(r.Context(), itemID, req.Name, req.Quantity)
	if err != nil {
		listServiceError(w, err, "item")
		return
	}

	writeJSON(w, http.StatusCreated, listItemJSONFromEntry(entry))
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

	entry, err := s.svc.UpdateItem(r.Context(), int32(id), req.Quantity, req.Done)
	if err != nil {
		listServiceError(w, err, "list item")
		return
	}

	writeJSON(w, http.StatusOK, listItemJSONFromEntry(entry))
}

func (s *Server) listRemoveItemHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		badRequest(w, "id is required")
		return
	}

	parsed, err := strconv.Atoi(id)
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	if err := s.svc.RemoveItem(r.Context(), int32(parsed)); err != nil {
		listServiceError(w, err, "list item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listFinishHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Finish(r.Context()); err != nil {
		internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// listServiceError maps service-layer sentinel errors onto API status codes.
func listServiceError(w http.ResponseWriter, err error, resource string) {
	switch {
	case errors.Is(err, service.ErrValidation):
		badRequest(w, err.Error())
	case errors.Is(err, service.ErrNotFound):
		notFound(w, resource)
	case errors.Is(err, service.ErrConflict):
		conflict(w, err.Error())
	default:
		internalError(w, err)
	}
}

// ---------------------------------------------------------------------------
// JSON representation helpers
// ---------------------------------------------------------------------------

type listItemJSON struct {
	ID         int32  `json:"id"`
	ItemID     int32  `json:"item_id"`
	ItemName   string `json:"item_name"`
	CategoryID int32  `json:"category_id"`
	Quantity   string `json:"quantity"`
	Done       bool   `json:"done"`
}

func listItemJSONFromEntry(entry service.ListEntry) listItemJSON {
	return listItemJSON{
		ID:         entry.ID,
		ItemID:     entry.ItemID,
		ItemName:   entry.Name,
		CategoryID: entry.CategoryID,
		Quantity:   entry.Quantity,
		Done:       entry.Done,
	}
}
