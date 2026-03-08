package server

import (
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/client"
)

func (s *Server) categoriesHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Categories []storeWithCategories
		Stores     []client.Store
	}

	bag := data{baseBag: s.newBag(r.Context())}

	apiClient := clientFromContext(r.Context())

	stores, err := apiClient.ListStores(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	categories, err := apiClient.ListCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Stores = stores
	bag.Categories = buildStoreHierarchy(stores, categories)

	renderHtml(w, http.StatusOK, "categories.gohtml", bag)
}

func (s *Server) categoryHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		baseBag
		Category client.CategoryDetail
		Items    []client.Item
		Stores   []client.Store
	}

	bag := data{baseBag: s.newBag(r.Context())}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	apiClient := clientFromContext(r.Context())

	bag.Category, err = apiClient.GetCategory(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Items = bag.Category.Items

	bag.Stores, err = apiClient.ListStores(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	renderHtml(w, http.StatusOK, "category.gohtml", bag)
}

func (s *Server) categoryAddHandler(w http.ResponseWriter, r *http.Request) {
	storeID, err := strconv.Atoi(r.FormValue("storeID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	apiClient := clientFromContext(r.Context())
	_, err = apiClient.CreateCategory(
		r.Context(),
		storeID,
		r.FormValue("name"),
		r.FormValue("description"),
	)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	s.sseServer.Publish(r.Context(), sseEventCategory, nil)

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (s *Server) categoryEditHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	storeID, err := strconv.Atoi(r.FormValue("storeID"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	apiClient := clientFromContext(r.Context())
	_, err = apiClient.UpdateCategory(
		r.Context(),
		id,
		storeID,
		r.FormValue("name"),
		r.FormValue("description"),
	)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	s.sseServer.Publish(r.Context(), sseEventCategory, nil)

	http.Redirect(w, r, "/categories", http.StatusFound)
}

func (s *Server) categoryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	apiClient := clientFromContext(r.Context())
	if err := apiClient.DeleteCategory(r.Context(), id); err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	s.sseServer.Publish(r.Context(), sseEventCategory, nil)

	http.Redirect(w, r, "/categories", http.StatusFound)
}

// buildStoreHierarchy groups a flat list of categories under their parent stores,
// producing the nested structure expected by the categories template.
func buildStoreHierarchy(stores []client.Store, categories []client.Category) []storeWithCategories {
	ret := make([]storeWithCategories, 0, len(stores))

	for _, store := range stores {
		node := storeWithCategories{Store: store}

		for _, cat := range categories {
			if cat.StoreID != store.ID {
				continue
			}
			node.Categories = append(node.Categories, categoryWithItems{Category: cat})
		}

		ret = append(ret, node)
	}

	return ret
}
