package server

import (
	"net/http"

	"github.com/taiidani/groceries/internal/db/models"
)

func (s *Server) categoriesHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Stores []models.Store
	}

	bag := data{baseBag: s.newBag(r.Context())}

	stores, err := s.db.ListStores(r.Context())
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

	bag.Category, err = s.db.GetCategory(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Items, err = s.db.ListItemsForCategory(r.Context(), bag.Category.ID)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Stores, err = s.db.ListStores(r.Context())
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

	_, err = s.db.CreateCategory(r.Context(), models.CreateCategoryParams{
		StoreID:     storeID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	})
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	s.sseServer.Publish(r.Context(), sseEventCategory, nil)

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

	_, err = s.db.UpdateCategory(r.Context(), models.UpdateCategoryParams{
		ID:          id,
		StoreID:     storeID,
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	})
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	s.sseServer.Publish(r.Context(), sseEventCategory, nil)

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (s *Server) categoryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := s.db.DeleteCategory(r.Context(), id); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	s.sseServer.Publish(r.Context(), sseEventCategory, nil)

	http.Redirect(w, r, "/categories", http.StatusFound)
}
