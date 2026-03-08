package server

import (
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/client"
)

type itemsBag struct {
	baseBag
	Stores []storeWithCategories
	Item   client.Item
}

func (s *Server) itemsHandler(w http.ResponseWriter, r *http.Request) {
	bag := itemsBag{baseBag: s.newBag(r.Context())}

	var err error
	bag.Stores, err = loadStoreHierarchy(r.Context(), storeHierarchyInput{})
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
		Categories []client.Category
		Item       client.Item
	}{baseBag: s.newBag(r.Context())}

	bag.Redirect = r.URL.Query().Get("redirect")

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	apiClient := clientFromContext(r.Context())

	categories, err := apiClient.ListCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
	bag.Categories = categories

	bag.Item, err = apiClient.GetItem(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	renderHtml(w, http.StatusOK, "item_edit.gohtml", bag)
}

func (s *Server) itemAddHandler(w http.ResponseWriter, r *http.Request) {
	categoryID, err := strconv.Atoi(r.FormValue("categoryID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	apiClient := clientFromContext(r.Context())
	_, err = apiClient.CreateItem(r.Context(), categoryID, r.FormValue("name"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	s.sseServer.Publish(r.Context(), sseEventList, nil)

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

	apiClient := clientFromContext(r.Context())

	_, err = apiClient.UpdateItem(r.Context(), id, categoryID, r.FormValue("name"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	existing, err := apiClient.GetItem(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if existing.List != nil {
		if err := apiClient.UpdateListItem(r.Context(), id, r.FormValue("quantity")); err != nil {
			errorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	s.sseServer.Publish(r.Context(), sseEventList, nil)

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

	apiClient := clientFromContext(r.Context())
	if err := apiClient.DeleteItem(r.Context(), id); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	s.sseServer.Publish(r.Context(), sseEventList, nil)

	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/items"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}
