package server

import (
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) categoriesHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Categories []storeWithCategories
		Stores     []models.Store
		Category   models.Category
	}

	bag := data{baseBag: s.newBag(r.Context())}

	var err error
	bag.Categories, err = loadStoreHierarchy(r.Context(), storeHierarchyInput{})
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Stores, err = models.LoadStores(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	template := "categories.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) categoryHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Category models.Category
		Stores   []models.Store
	}

	bag := data{baseBag: s.newBag(r.Context())}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	bag.Category, err = models.GetCategory(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Stores, err = models.LoadStores(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	template := "category.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) categoryAddHandler(w http.ResponseWriter, r *http.Request) {
	storeID, err := strconv.Atoi(r.FormValue("storeID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	newCategory := models.Category{
		Name:        r.FormValue("name"),
		StoreID:     storeID,
		Description: r.FormValue("description"),
	}

	// Validate inputs
	if err := newCategory.Validate(r.Context()); err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	// Add the new category
	err = models.AddCategory(r.Context(), newCategory)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventCategory, nil)

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (s *Server) categoryEditHandler(w http.ResponseWriter, r *http.Request) {
	storeID, err := strconv.Atoi(r.FormValue("storeID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	newCategory := models.Category{
		ID:          r.FormValue("id"),
		Name:        r.FormValue("name"),
		StoreID:     storeID,
		Description: r.FormValue("description"),
	}

	// Validate inputs
	if err := newCategory.Validate(r.Context()); err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	// Add the new category
	err = models.EditCategory(r.Context(), newCategory)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventCategory, nil)

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (s *Server) categoryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteCategory(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.Publish(r.Context(), sseEventCategory, nil)

	http.Redirect(w, r, "/categories", http.StatusFound)
}
