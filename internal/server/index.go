package server

import (
	"net/http"
	"sort"

	"github.com/taiidani/groceries/internal/service"
)

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	type itemWithCategory struct {
		Category string
		Name     string
	}

	type indexBag struct {
		baseBag
		Items []itemWithCategory
	}

	bag := indexBag{baseBag: s.newBag(r.Context())}

	inList := false
	items, err := s.svc.ListItems(r.Context(), service.ItemFilters{InList: &inList})
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	for _, item := range items {
		bag.Items = append(bag.Items, itemWithCategory{
			Category: item.CategoryName,
			Name:     item.Name,
		})
	}

	sort.Slice(bag.Items, func(i, j int) bool {
		return bag.Items[i].Name < bag.Items[j].Name
	})

	renderHtml(w, http.StatusOK, "index.gohtml", bag)
}

func (s *Server) indexListHandler(w http.ResponseWriter, r *http.Request) {
	type indexListBag struct {
		baseBag
		Total     int
		TotalDone int
		List      []storeWithCategories
	}

	bag := indexListBag{baseBag: s.newBag(r.Context())}

	var err error
	bag.List, err = s.loadStoreHierarchy(r.Context(), storeHierarchyInput{
		OnlyListItems:         true,
		ExcludeDoneItems:      true,
		ExcludeEmptyGroupings: true,
	})
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	list, err := s.svc.GetList(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	bag.Total = list.Total
	bag.TotalDone = list.TotalDone

	renderHtml(w, http.StatusOK, "index_list.gohtml", bag)
}

func (s *Server) indexCartHandler(w http.ResponseWriter, r *http.Request) {
	type indexCartBag struct {
		baseBag
		DoneCategories []categoryWithItems
	}

	bag := indexCartBag{baseBag: s.newBag(r.Context())}

	categories, err := s.svc.ListCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	list, err := s.svc.GetList(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	doneByCategory := make(map[int32][]service.HierarchyItem)
	for _, entry := range list.Items {
		if !entry.Done {
			continue
		}
		doneByCategory[entry.CategoryID] = append(doneByCategory[entry.CategoryID], service.HierarchyItem{
			ID:         entry.ItemID,
			CategoryID: entry.CategoryID,
			Name:       entry.Name,
			List: &service.HierarchyListEntry{
				ID:       entry.ID,
				Quantity: entry.Quantity,
				Done:     entry.Done,
			},
		})
	}

	for _, cat := range categories {
		done := doneByCategory[cat.ID]
		if len(done) > 0 {
			bag.DoneCategories = append(bag.DoneCategories, categoryWithItems{
				Category: cat,
				Items:    done,
			})
		}
	}

	renderHtml(w, http.StatusOK, "index_cart.gohtml", bag)
}
