package server

import (
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/client"
	"github.com/taiidani/groceries/internal/models"
)

func (s *Server) storesHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Stores []client.Store
		Store  client.Store
	}

	bag := data{baseBag: s.newBag(r.Context())}

	apiClient := clientFromContext(r.Context())
	stores, err := apiClient.ListStores(r.Context())
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
		Store      client.Store
		Categories []client.Category
	}

	bag := data{baseBag: s.newBag(r.Context())}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	apiClient := clientFromContext(r.Context())
	store, err := apiClient.GetStore(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Store = store
	bag.Categories = store.Categories

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
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	apiClient := clientFromContext(r.Context())
	_, err = apiClient.UpdateStore(r.Context(), id, r.FormValue("name"))
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

	apiClient := clientFromContext(r.Context())
	if err := apiClient.DeleteStore(r.Context(), id); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/stores", http.StatusFound)
}

type storeWithCategories struct {
	models.Store
	Categories []categoryWithItems
}

type categoryWithItems struct {
	models.Category
	Items []models.Item
}
