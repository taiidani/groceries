package server

import (
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) storesHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Stores []models.Store
		Store  models.Store
	}

	bag := data{baseBag: s.newBag(r.Context())}

	stores, err := models.LoadStores(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Stores = stores

	template := "stores.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) storeHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Store models.Store
	}

	bag := data{baseBag: s.newBag(r.Context())}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	bag.Store, err = models.GetStore(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	template := "store.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) storeAddHandler(w http.ResponseWriter, r *http.Request) {
	newStore := models.Store{
		Name: r.FormValue("name"),
	}

	// Validate inputs
	if err := newStore.Validate(r.Context()); err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	// Add the new Store
	err := models.AddStore(r.Context(), newStore)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/stores", http.StatusFound)
}

func (s *Server) storeEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	newStore := models.Store{
		ID:   id,
		Name: r.FormValue("name"),
	}

	// Validate inputs
	if err := newStore.Validate(r.Context()); err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	// Add the new Store
	err = models.EditStore(r.Context(), newStore)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/stores", http.StatusFound)
}

func (s *Server) storeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	err = models.DeleteStore(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/stores", http.StatusFound)
}
