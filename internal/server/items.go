package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/db/models"
	"github.com/taiidani/groceries/internal/service"
)

type itemsBag struct {
	baseBag
	Stores []storeWithCategories
	Item   service.Item
}

// editItem is the view model for the item edit form, keeping the template's
// .Item.List.Quantity shape while sourcing data from the service layer.
type editItem struct {
	service.Item
	List *editItemList
}

type editItemList struct {
	ID       int32
	Quantity string
	Done     bool
}

func (s *Server) itemsHandler(w http.ResponseWriter, r *http.Request) {
	bag := itemsBag{baseBag: s.newBag(r.Context())}

	var err error
	bag.Stores, err = s.loadStoreHierarchy(r.Context(), storeHierarchyInput{})
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	renderHtml(w, http.StatusOK, "items.gohtml", bag)
}

func (s *Server) itemHandler(w http.ResponseWriter, r *http.Request) {
	bag := struct {
		baseBag
		Redirect   string
		Categories []models.Category
		Item       editItem
	}{baseBag: s.newBag(r.Context())}

	bag.Redirect = r.URL.Query().Get("redirect")

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	categories, err := s.svc.ListCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
	bag.Categories = categories

	item, err := s.svc.GetItem(r.Context(), int32(id))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
	bag.Item = editItem{Item: item}
	if item.OnList {
		bag.Item.List = &editItemList{
			ID:       item.ListID,
			Quantity: item.ListQuantity,
			Done:     item.ListDone,
		}
	}

	renderHtml(w, http.StatusOK, "item_edit.gohtml", bag)
}

func (s *Server) itemAddHandler(w http.ResponseWriter, r *http.Request) {
	categoryID, err := strconv.Atoi(r.FormValue("categoryID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if _, err := s.svc.CreateItem(r.Context(), int32(categoryID), r.FormValue("name")); err != nil {
		errorResponse(w, r, itemErrorStatus(err), err)
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/items"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) itemEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	categoryID, err := strconv.Atoi(r.FormValue("categoryID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	// Only pass the quantity when the item is on the shopping list; the
	// service applies the item update and the list-quantity update in a
	// single transaction (previously two separate API calls).
	var quantity *string
	if existing, err := s.svc.GetItem(r.Context(), int32(id)); err == nil && existing.OnList {
		q := r.FormValue("quantity")
		quantity = &q
	}

	if _, err := s.svc.UpdateCatalogItem(r.Context(), int32(id), r.FormValue("name"), int32(categoryID), quantity); err != nil {
		errorResponse(w, r, itemErrorStatus(err), err)
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/items"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Server) itemDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := s.svc.DeleteItem(r.Context(), int32(id)); err != nil {
		errorResponse(w, r, itemErrorStatus(err), err)
		return
	}

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/items"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

// itemErrorStatus maps service-layer sentinel errors onto HTTP status codes
// for the web transport.
func itemErrorStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
