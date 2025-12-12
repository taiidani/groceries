package server

import (
	"context"
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
		Store      models.Store
		Categories []models.Category
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

	bag.Categories, err = bag.Store.Categories(r.Context())
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

type storeWithCategories struct {
	models.Store
	Categories []categoryWithItems
}

type categoryWithItems struct {
	models.Category
	Items []models.Item
}

type storeHierarchyInput struct {
	ExcludeEmptyGroupings bool
	ExcludeDoneItems      bool
	OnlyListItems         bool
}

func loadStoreHierarchy(ctx context.Context, input storeHierarchyInput) ([]storeWithCategories, error) {
	ret := []storeWithCategories{}

	stores, err := models.LoadStores(ctx)
	if err != nil {
		return ret, err
	}

	categories, err := models.LoadCategories(ctx)
	if err != nil {
		return ret, err
	}

	var items []models.Item
	if input.OnlyListItems {
		items, err = models.LoadList(ctx)
	} else {
		items, err = models.LoadItems(ctx)
	}
	if err != nil {
		return ret, err
	}

	for _, store := range stores {
		addStore := storeWithCategories{Store: store}

		for _, cat := range categories {
			if cat.StoreID != addStore.Store.ID {
				continue
			}

			addItem := []models.Item{}
			for _, item := range items {
				if item.CategoryID != cat.ID {
					continue
				}
				if input.ExcludeDoneItems && item.List.Done {
					continue
				}

				addItem = append(addItem, item)
			}

			if !input.ExcludeEmptyGroupings || len(addItem) > 0 {
				addStore.Categories = append(addStore.Categories, categoryWithItems{
					Category: cat,
					Items:    addItem,
				})
			}
		}

		if !input.ExcludeEmptyGroupings || len(addStore.Categories) > 0 {
			ret = append(ret, addStore)
		}
	}

	return ret, nil
}
