package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/service"
)

// ---------------------------------------------------------------------------
// JSON representation helpers
// ---------------------------------------------------------------------------

// itemJSON replicates the wire format previously produced by
// internal/models.Item.MarshalJSON: category_name plus an optional embedded
// list object.
type itemJSON struct {
	ID           int32         `json:"id"`
	CategoryID   int32         `json:"category_id"`
	CategoryName string        `json:"category_name"`
	Name         string        `json:"name"`
	List         *itemListJSON `json:"list"`
}

type itemListJSON struct {
	ID         int32  `json:"id"`
	ItemID     int32  `json:"item_id"`
	CategoryID string `json:"category_id"`
	Quantity   string `json:"quantity"`
	Done       bool   `json:"done"`
	Name       string `json:"name"`
}

// itemJSONFromService converts a service-layer Item to its wire format. When
// full is false the list object is populated from the lightweight summary
// fields (ID/quantity/done only), matching the previous list-endpoint shape.
func itemJSONFromService(item service.Item, full bool) itemJSON {
	ret := itemJSON{
		ID:           item.ID,
		CategoryID:   item.CategoryID,
		CategoryName: item.CategoryName,
		Name:         item.Name,
	}
	if item.OnList {
		entry := &itemListJSON{
			ID:       item.ListID,
			Quantity: item.ListQuantity,
			Done:     item.ListDone,
		}
		if full {
			entry.ItemID = item.ID
			entry.CategoryID = strconv.Itoa(int(item.CategoryID))
			entry.Name = item.Name
		}
		ret.List = entry
	}
	return ret
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func (s *Server) itemsListHandler(w http.ResponseWriter, r *http.Request) {
	var filters service.ItemFilters

	q := r.URL.Query()
	if rawID := q.Get("category_id"); rawID != "" {
		categoryID, err := strconv.Atoi(rawID)
		if err != nil {
			badRequest(w, "category_id must be an integer")
			return
		}
		id := int32(categoryID)
		filters.CategoryID = &id
	}

	if rawInList := q.Get("in_list"); rawInList != "" {
		inList, err := strconv.ParseBool(rawInList)
		if err != nil {
			badRequest(w, "in_list must be a boolean")
			return
		}
		filters.InList = &inList
	}

	items, err := s.svc.ListItems(r.Context(), filters)
	if err != nil {
		internalError(w, err)
		return
	}

	ret := make([]itemJSON, 0, len(items))
	for _, item := range items {
		ret = append(ret, itemJSONFromService(item, false))
	}

	writeJSON(w, http.StatusOK, ret)
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

	item, err := s.svc.CreateItem(r.Context(), int32(req.CategoryID), req.Name)
	if err != nil {
		itemServiceError(w, err, "item")
		return
	}

	writeJSON(w, http.StatusCreated, itemJSONFromService(item, true))
}

func (s *Server) itemsGetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	item, err := s.svc.GetItem(r.Context(), int32(id))
	if err != nil {
		itemServiceError(w, err, "item")
		return
	}

	writeJSON(w, http.StatusOK, itemJSONFromService(item, true))
}

func (s *Server) itemsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
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

	// The API PUT does not carry list quantity; pass nil to leave it alone.
	item, err := s.svc.UpdateCatalogItem(r.Context(), int32(id), req.Name, int32(req.CategoryID), nil)
	if err != nil {
		itemServiceError(w, err, "item")
		return
	}

	writeJSON(w, http.StatusOK, itemJSONFromService(item, true))
}

func (s *Server) itemsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		badRequest(w, "id must be an integer")
		return
	}

	if err := s.svc.DeleteItem(r.Context(), int32(id)); err != nil {
		itemServiceError(w, err, "item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// itemServiceError maps service-layer sentinel errors onto API status codes.
func itemServiceError(w http.ResponseWriter, err error, resource string) {
	listServiceError(w, err, resource)
}
