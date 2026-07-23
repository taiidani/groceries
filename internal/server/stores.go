package server

import (
	"net/http"

	"github.com/taiidani/groceries/internal/client"
	"github.com/taiidani/groceries/internal/db/models"
)

func (s *Server) storesHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Stores []models.Store
		Store  models.Store
	}

	bag := data{baseBag: s.newBag(r.Context())}

	stores, err := s.db.ListStores(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Stores = stores

	renderHtml(w, http.StatusOK, "stores.gohtml", bag)
}

func (s *Server) storeHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Store      models.Store
		Categories []models.Category
	}

	bag := data{baseBag: s.newBag(r.Context())}

	id, err := parseId(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	store, err := s.db.GetStore(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	categories, err := s.db.ListCategoriesForStore(r.Context(), store.ID)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Store = store
	bag.Categories = categories

	renderHtml(w, http.StatusOK, "store.gohtml", bag)
}

func (s *Server) storeAddHandler(w http.ResponseWriter, r *http.Request) {
	apiClient := clientFromContext(r.Context())
	_, err := apiClient.CreateStore(r.Context(), r.FormValue("name"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/stores", http.StatusFound)
}

func (s *Server) storeEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	_, err = s.db.UpdateStore(r.Context(), models.UpdateStoreParams{
		ID:   id,
		Name: r.FormValue("name"),
	})
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/stores", http.StatusFound)
}

func (s *Server) storeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseId(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := s.db.DeleteStore(r.Context(), id); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/stores", http.StatusFound)
}

type storeWithCategories struct {
	client.Store
	Categories []categoryWithItems
}

type categoryWithItems struct {
	client.Category
	Items []client.Item
}
