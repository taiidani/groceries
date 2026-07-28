package server

import (
	"net/http"

	"github.com/taiidani/groceries/internal/db/models"
	"github.com/taiidani/groceries/internal/service"
)

func (s *Server) storesHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Stores []models.Store
		Store  models.Store
	}

	bag := data{baseBag: s.newBag(r.Context())}

	stores, err := s.svc.ListStores(r.Context())
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

	detail, err := s.svc.GetStore(r.Context(), id)
	if err != nil {
		categoryStoreErrorResponse(w, r, err)
		return
	}

	bag.Store = detail.Store
	bag.Categories = detail.Categories

	renderHtml(w, http.StatusOK, "store.gohtml", bag)
}

func (s *Server) storeAddHandler(w http.ResponseWriter, r *http.Request) {
	_, err := s.svc.CreateStore(r.Context(), r.FormValue("name"))
	if err != nil {
		categoryStoreErrorResponse(w, r, err)
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

	_, err = s.svc.UpdateStore(r.Context(), id, r.FormValue("name"))
	if err != nil {
		categoryStoreErrorResponse(w, r, err)
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

	if err := s.svc.DeleteStore(r.Context(), id); err != nil {
		categoryStoreErrorResponse(w, r, err)
		return
	}

	http.Redirect(w, r, "/stores", http.StatusFound)
}

type storeWithCategories = service.StoreWithCategories

type categoryWithItems = service.CategoryWithItems
