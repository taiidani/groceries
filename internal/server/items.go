package server

import (
	"net/http"
	"strconv"

	"github.com/taiidani/groceries/internal/models"
)

type itemsBag struct {
	baseBag
	Categories []categoryWithItems
}

func (s *Server) itemsHandler(w http.ResponseWriter, r *http.Request) {
	bag := itemsBag{baseBag: s.newBag(r.Context())}

	categories, err := models.LoadCategories(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	items, err := models.LoadItems(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	for _, cat := range categories {
		add := []models.Item{}
		for _, item := range items {
			if item.CategoryID == cat.ID {
				add = append(add, item)
			}
		}

		if len(add) > 0 {
			bag.Categories = append(bag.Categories, categoryWithItems{
				Category: models.Category{
					ID:          cat.ID,
					Description: cat.Description,
					Name:        cat.Name,
				},
				Items: add,
			})
		}
	}

	template := "items.gohtml"
	renderHtml(w, http.StatusOK, template, bag)
}

func (s *Server) listDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteFromList(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.announce(sseEventList)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) itemBagHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	err = models.AddExistingItem(r.Context(), id)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.announce(sseEventBag)

	http.Redirect(w, r, "/items", http.StatusFound)
}

func (s *Server) itemDeleteHandler(w http.ResponseWriter, r *http.Request) {
	err := models.DeleteItem(r.Context(), r.FormValue("id"))
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/items", http.StatusFound)
}

func (s *Server) itemDoneHandler(w http.ResponseWriter, r *http.Request) {
	err := models.MarkItemDone(r.Context(), r.FormValue("id"), true)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.announce(sseEventList, sseEventCart)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) itemUnDoneHandler(w http.ResponseWriter, r *http.Request) {
	err := models.MarkItemDone(r.Context(), r.FormValue("id"), false)
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.announce(sseEventList, sseEventCart)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) finishHandler(w http.ResponseWriter, r *http.Request) {
	err := models.FinishShopping(r.Context())
	if err != nil {
		errorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	// Broadcast the change
	s.sseServer.announce(sseEventCart)

	http.Redirect(w, r, "/", http.StatusFound)
}
