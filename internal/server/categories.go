package server

import (
	"errors"
	"net/http"

	"github.com/taiidani/groceries/internal/db/models"
	"github.com/taiidani/groceries/internal/service"
)

func (s *Server) categoriesHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Stores []models.Store
	}

	bag := data{baseBag: s.newBag(r.Context())}

	stores, err := s.svc.ListStores(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Stores = stores
	renderHtml(w, http.StatusOK, "categories.gohtml", bag)
}

func (s *Server) partialCategoriesListForStoreHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	categories, err := s.db.ListCategoriesForStoreWithItemCount(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	renderHtml(w, http.StatusOK, "_categories_list_for_store.gohtml", categories)
}

func (s *Server) categoryHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Category models.Category
		Items    []models.Item
		Stores   []models.Store
	}

	bag := data{baseBag: s.newBag(r.Context())}

	id, err := parseId(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	detail, err := s.svc.GetCategory(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) {
			status = http.StatusNotFound
		}
		errorResponse(w, r, status, err)
		return
	}

	bag.Category = detail.Category
	bag.Items = detail.Items

	bag.Stores, err = s.svc.ListStores(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	renderHtml(w, http.StatusOK, "category.gohtml", bag)
}

func (s *Server) categoryAddHandler(w http.ResponseWriter, r *http.Request) {
	storeID, err := parseId(r.FormValue("storeID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	_, err = s.svc.CreateCategory(r.Context(), storeID, r.FormValue("name"), r.FormValue("description"))
	if err != nil {
		categoryStoreErrorResponse(w, r, err)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (s *Server) categoryEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	storeID, err := parseId(r.FormValue("storeID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	_, err = s.svc.UpdateCategory(r.Context(), id, storeID, r.FormValue("name"), r.FormValue("description"))
	if err != nil {
		categoryStoreErrorResponse(w, r, err)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (s *Server) categoryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := s.svc.DeleteCategory(r.Context(), id); err != nil {
		categoryStoreErrorResponse(w, r, err)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusFound)
}

// categoryStoreErrorResponse maps service sentinel errors onto HTTP status
// codes for the web transport.
func categoryStoreErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		errorResponse(w, r, http.StatusBadRequest, err)
	case errors.Is(err, service.ErrNotFound):
		errorResponse(w, r, http.StatusNotFound, err)
	case errors.Is(err, service.ErrConflict):
		errorResponse(w, r, http.StatusConflict, err)
	default:
		errorResponse(w, r, http.StatusInternalServerError, err)
	}
}
